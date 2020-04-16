package quadgo

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

var quadClient *QuadrilleClient

func IsPortOpen(host string, port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			return IsPortOpen(host, port, timeout)
		}
		return false
	}
	defer conn.Close()
	return true
}

var ln *net.TCPListener

func startMockService() {
	s, err := net.ResolveTCPAddr("tcp", "localhost:5679")
	if err != nil {
		return
	}

	ln, err = net.ListenTCP("tcp", s)
	if err != nil {
		return
	}
	go func() {
		defer ln.Close()
		for {
			c, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				break
			}
			go func(c net.Conn) {
				defer c.Close()
				for {
					cmdLine, err := bufio.NewReader(c).ReadString('\n')
					if err != nil {
						fmt.Println("handleConnection", err)
						break
					}
					cmdLine = strings.TrimSpace(cmdLine)
					cmdParts := strings.Split(cmdLine, "::")
					if strings.Index(cmdParts[1], "insert") == 0 {
						c.Write([]byte(fmt.Sprintf("%s::%s\n", cmdParts[0], "ok")))
					} else if strings.Index(cmdParts[1], "get") == 0 {
						c.Write([]byte(fmt.Sprintf("%s::%s\n", cmdParts[0], `{"data":"loc2","lat":17,"lon":78}`)))
					} else if strings.Index(cmdParts[1], "neighbors") == 0 {
						c.Write([]byte(fmt.Sprintf("%s::%s\n", cmdParts[0], `[{"data":"loc2","lat":12.99478,"lon":77.542687}]`)))
					} else if strings.Index(cmdParts[1], "members") == 0 {
						c.Write([]byte(fmt.Sprintf("%s::%s\n", cmdParts[0], `[{"id":"node0","addr":"127.0.0.1:5679"}]`)))
					} else if strings.Index(cmdParts[1], "isleader") == 0 {
						c.Write([]byte(fmt.Sprintf("%s::%s\n", cmdParts[0], "true")))
					}
				}

			}(c)
		}
	}()
}

func init() {
	isPortAlreadyInUse := IsPortOpen("localhost", 5679, time.Second*1)
	if isPortAlreadyInUse {
		return
	}
	startMockService()
	fmt.Println("mock service started")
	quadClient = NewClient("localhost:5679")
}

func TestDriver(t *testing.T) {
	if quadClient == nil {
		fmt.Println("skipping quadgo tests")
		return
	}
	locationID, latitude, longitude := "loc123", 17.0, 78.0
	testInsert(t, locationID, latitude, longitude)
	testGet(t, locationID, latitude, longitude)
	testGetNeighbors(t, latitude, longitude)

	quadClient.Close()
	ln.Close()
}

func testInsert(t *testing.T, locationID string, lat, lon float64) {
	err := quadClient.Insert(locationID, lat, lon, map[string]interface{}{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}

func testGet(t *testing.T, locationID string, lat, lon float64) {
	loc, err := quadClient.Get(locationID)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if loc.Latitude != lat || loc.Longitude != lon {
		t.Fatalf("Expected: %f,%f Got:%f,%f", lat, lon, loc.Latitude, loc.Longitude)
	}
}

func testGetNeighbors(t *testing.T, lat, lon float64) {
	neighbors, _ := quadClient.Nearby(lat, lon, 1000, 10)
	if len(neighbors) == 0 {
		t.Fatalf("Didn't get neighbors")
	}
}

//func TestBulkWrite(t *testing.T) {
//	bulk := NewBulkWrite()
//	bulk.Add(NewInsertOperation("loc123", 12, 77, map[string]interface{}{}))
//	bulk.Add(NewUpdateOperation("loc123", 17, 78, map[string]interface{}{"reg_no": "KA03NB5352"}))
//	client := NewClient("localhost:5679")
//	client.ExecuteBulk(bulk)
//}
