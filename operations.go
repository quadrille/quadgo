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

func (q *QuadrilleClient) ExecuteBulk(bulk *bulkWrite) (err error) {
	s, err := bulk.GetStr()
	if err != nil {
		return
	}
	_, err = q.getResponse(s)
	return
}

type bulkWrite struct {
	operations []bulkWriteOperation
}

func NewBulkWrite() *bulkWrite {
	return &bulkWrite{operations: make([]bulkWriteOperation, 0)}
}

func (b *bulkWrite) Add(operation *bulkWriteOperation) {
	b.operations = append(b.operations, *operation)
}

func (b *bulkWrite) GetStr() (s string, err error) {
	if len(b.operations) == 0 {
		err = ErrEmptyBulkWrite
	}
	dataBytes, err := json.Marshal(b.operations)
	s = fmt.Sprintf("bulkwrite %s\n", string(dataBytes))
	return
}

type bulkWriteOperation struct {
	Op         string                 `json:"op"`
	LocationID string                 `json:"location_id"`
	Lat        float64                `json:"lat"`
	Lon        float64                `json:"lon"`
	Data       map[string]interface{} `json:"data"`
}

func NewInsertOperation(locationID string, lat, lon float64, data map[string]interface{}) *bulkWriteOperation {
	return &bulkWriteOperation{
		Op:         "insert",
		LocationID: locationID,
		Lat:        lat,
		Lon:        lon,
		Data:       data,
	}
}

func NewUpdateOperation(locationID string, lat, lon float64, data map[string]interface{}) *bulkWriteOperation {
	return &bulkWriteOperation{
		Op:         "update",
		LocationID: locationID,
		Lat:        lat,
		Lon:        lon,
		Data:       data,
	}
}

func NewUpdateLocOperation(locationID string, lat, lon float64) *bulkWriteOperation {
	return &bulkWriteOperation{
		Op:         "updateloc",
		LocationID: locationID,
		Lat:        lat,
		Lon:        lon,
	}
}

func NewUpdateDataOperation(locationID string, data map[string]interface{}) *bulkWriteOperation {
	return &bulkWriteOperation{
		Op:         "update",
		LocationID: locationID,
		Data:       data,
	}
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
