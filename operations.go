package quadgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (q *QuadrilleClient) getResponse(query string) (string, error) {
	return func(responseCh chan string, reqID uint64, err error) (string, error) {
		if err != nil {
			return "", err
		}
		select {
		case responseStr := <-responseCh:
			if strings.HasPrefix(responseStr, "ERROR:") {
				errMsgStart := strings.Index(responseStr, "ERROR:") + 6
				return "", errors.New(responseStr[errMsgStart:])
			} else {
				return responseStr, nil
			}
		case <-time.After(time.Second * 1):
			q.removeReqFromMap(reqID)
			fmt.Printf("query #%d timed-out\n", int(reqID))
			return "", errors.New("timeout")
		}
	}(q.submitQuery(query))
}

func (q *QuadrilleClient) Get(locationID string) (*Location, error) {
	responseStr, err := q.getResponse(fmt.Sprintf("get %s", locationID))
	if err != nil {
		return nil, err
	}
	var location Location
	err = json.Unmarshal([]byte(responseStr), &location)
	if err != nil {
		return nil, err
	}
	return &location, nil
}

func (q *QuadrilleClient) Delete(locationID string) (err error) {
	_, err = q.getResponse(fmt.Sprintf("del %s", locationID))
	return
}

func (q *QuadrilleClient) Nearby(latitude, longitude float64, radiusInMeters, limit int) (neighbors []NeighborResult, err error) {
	neighborsStr, err := q.getResponse(fmt.Sprintf("neighbors %f,%f %d %d", latitude, longitude, radiusInMeters, limit))
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(neighborsStr), &neighbors)
	return
}

func (q *QuadrilleClient) Insert(locationID string, latitude, longitude float64, data map[string]interface{}) (err error) {
	dataBytes, err := json.Marshal(data)
	_, err = q.getResponse(fmt.Sprintf("insert %s %f,%f %s", locationID, latitude, longitude, string(dataBytes)))
	return
}

func (q *QuadrilleClient) Update(locationID string, latitude, longitude float64, data map[string]interface{}) (err error) {
	dataBytes, err := json.Marshal(data)
	_, err = q.getResponse(fmt.Sprintf("update %s %f,%f %s", locationID, latitude, longitude, string(dataBytes)))
	return
}

func (q *QuadrilleClient) UpdateLocation(locationID string, latitude, longitude float64) (err error) {
	_, err = q.getResponse(fmt.Sprintf("updateloc %s %f,%f", locationID, latitude, longitude))
	return
}

func (q *QuadrilleClient) UpdateData(locationID string, data map[string]interface{}) (err error) {
	dataBytes, err := json.Marshal(data)
	_, err = q.getResponse(fmt.Sprintf("updatedata %s %s", locationID, string(dataBytes)))
	return
}

//func (q *QuadrilleClient) IsLeader() (isLeader bool, err error) {
//	responseStr, err := q.getResponse(fmt.Sprintf("isleader")))
//	isLeader = false
//	if err != nil {
//		return
//	}
//	isLeader, err = strconv.ParseBool(responseStr)
//	if err != nil {
//		return
//	}
//	return
//}
