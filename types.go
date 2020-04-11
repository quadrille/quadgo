package quadgo

type LocationID string

type Location struct {
	ID        LocationID  `json:"location_id""`
	Latitude  float64     `json:"lat"`
	Longitude float64     `json:"lon"`
	Data      interface{} `json:"data"`
}

type NeighborResult struct {
	Location
	Distance float64 `json:"distance"`
}
