package quadgo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Needs to be changed to a dynamic size later
//Probably have a min and max size and grow and shrink based in query rate
const connPoolSize = 100

var q *QuadrilleClient
var once sync.Once

type querySequencer struct {
	id uint64
	mu sync.Mutex
}

func (qs *querySequencer) generateID() uint64 {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.id++
	return qs.id
}

type QuadrilleClient struct {
	members          []string
	connPoolSize     int
	connPool         []net.Conn
	connPoolResetMtx sync.RWMutex
	querySeq         querySequencer
	respChannelMap   map[uint64]chan string
	respChnMtx       sync.Mutex
}

func NewClient(addr string) *QuadrilleClient {
	once.Do(func() {
		members := strings.Split(addr, ",")
		members = getMembers(members)
		q = &QuadrilleClient{
			members:        members,
			connPoolSize:   connPoolSize,
			querySeq:       querySequencer{},
			respChannelMap: map[uint64]chan string{},
		}
		leaderAddr, err := q.getLeader()
		if err != nil {
			log.Fatalln("unable to find any leader node")
		}
		q.connPool = prepareConnectionPool(leaderAddr)
		go q.responseNotifier()
	})
	return q
}

func prepareConnectionPool(addr string) []net.Conn {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Fatalln("ResolveTCPAddr:", err.Error())
	}
	connArr := make([]net.Conn, 0)
	for i := 0; i < connPoolSize; i++ {
		conn, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			log.Fatalln("DialTCP:", err.Error())
		}
		connArr = append(connArr, conn)
	}
	return connArr
}

func isLeader(addr string) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return false
	}
	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	defer conn.Close()
	if err != nil {
		return false
	}
	_, err = fmt.Fprintln(conn, "1::isleader")
	if err != nil {
		return false
	}
	message, err := bufio.NewReader(conn).ReadString('\n')
	message = strings.TrimSpace(message)
	if err != nil {
		return false
	}
	boolResp, err := strconv.ParseBool(strings.Split(message, "::")[1])
	if err != nil {
		return false
	}
	return boolResp
}

func getMembers(addrs []string) []string {
	for _, addr := range addrs {
		members := func(addr string) []string {
			type Member struct {
				ID   string `json:"id"`
				Addr string `json:"addr"`
			}
			tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
			if err != nil {
				return []string{}
			}
			conn, err := net.DialTCP("tcp4", nil, tcpAddr)
			if err != nil {
				return []string{}
			}
			defer conn.Close()
			_, err = fmt.Fprintln(conn, "1::members")
			if err != nil {
				return []string{}
			}
			message, err := bufio.NewReader(conn).ReadString('\n')
			message = strings.TrimSpace(message)
			message = strings.Split(message, "::")[1]
			if err != nil {
				return []string{}
			}
			var members []Member
			err = json.Unmarshal([]byte(message), &members)
			if err != nil {
				fmt.Println(err)
				return []string{}
			}
			tcpPort := strings.Split(addr, ":")[1]
			memberAddrs := make([]string, 0)
			for _, member := range members {
				memberAddrs = append(memberAddrs, strings.Split(member.Addr, ":")[0]+":"+tcpPort)
			}
			return memberAddrs
		}(addr)

		if len(members) > 0 {
			return members
		}
	}
	return []string{}
}

func (q *QuadrilleClient) getLeader() (string, error) {
	leaderSrchCh := make(chan string)
	for _, member := range q.members {
		go func() {
			flag := isLeader(member)
			if flag {
				leaderSrchCh <- member
			} else {
				leaderSrchCh <- ""
			}
		}()
	}
	leaderAddr := ""
	for i := 0; i < len(q.members); i++ {
		addr := <-leaderSrchCh
		if addr != "" {
			leaderAddr = addr
		}
	}
	if leaderAddr == "" {
		return "", NoLeaderFoundErr
	}
	return leaderAddr, nil
}

func (q *QuadrilleClient) closeConnections() {
	for _, conn := range q.connPool {
		conn.Close()
	}
}

func (q *QuadrilleClient) Close() {
	q.connPoolResetMtx.RLock()
	defer q.connPoolResetMtx.RUnlock()
	q.closeConnections()
}

func (q *QuadrilleClient) resetConnPool() {
	//Get current leader
	leader := ""
	for i := 0; i < 4; i++ {
		fmt.Printf("connect attempt %d\n", i)
		time.Sleep(time.Duration(float64(time.Second) * math.Pow(2, float64(i))))
		leaderTmp, err := q.getLeader()
		fmt.Printf("Leader:%s\n", leader)
		if err == nil {
			leader = leaderTmp
			break
		}
	}
	if leader != "" {
		q.connPoolResetMtx.Lock()
		defer q.connPoolResetMtx.Unlock()
		q.connPool = prepareConnectionPool(leader)
		go q.responseNotifier()
	} else {
		log.Fatalln("Unable to find a Quadrille leader")
	}
}

func (q *QuadrilleClient) submitQuery(line string) (responseCh chan string, reqID uint64, err error) {
	responseCh = make(chan string)
	reqID = q.querySeq.generateID()
	connIndex := reqID % connPoolSize

	q.connPoolResetMtx.RLock()
	conn := q.connPool[connIndex]
	q.connPoolResetMtx.RUnlock()

	q.respChnMtx.Lock()
	q.respChannelMap[reqID] = responseCh
	q.respChnMtx.Unlock()
	//fmt.Println("line->", line)
	//fmt.Printf("%d\n", reqID)
	_, err = fmt.Fprintf(conn, "%d::%s\n", reqID, line)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func (q *QuadrilleClient) responseNotifier() {
	var wg sync.WaitGroup
	for i := 0; i < connPoolSize; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				q.connPoolResetMtx.RLock()
				conn := q.connPool[i]
				q.connPoolResetMtx.RUnlock()
				message, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Println(err)
					q.closeConnections()
					break
				}
				//fmt.Println(message)
				msgParts := strings.Split(message, "::")
				id, _ := strconv.Atoi(msgParts[0])
				q.respChnMtx.Lock()
				responseCh, ok := q.respChannelMap[uint64(id)]
				q.respChnMtx.Unlock()
				//fmt.Println("Got Notification", responseCh, ok, reqChannelMap)
				if !ok {
					continue
				}
				//fmt.Print("->: " + message)
				responseCh <- msgParts[1]
				q.respChnMtx.Lock()
				delete(q.respChannelMap, uint64(id))
				q.respChnMtx.Unlock()
			}
		}(i)
	}
	wg.Wait()
	q.resetConnPool()
}

func (q *QuadrilleClient) removeReqFromMap(reqID uint64) {
	q.respChnMtx.Lock()
	defer q.respChnMtx.Unlock()
	delete(q.respChannelMap, reqID)
}
