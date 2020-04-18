## quadgo

###### The official QuadrilleDB Go Driver 

[![Travis Build](https://api.travis-ci.org/quadrille/quadgo.svg)](https://api.travis-ci.org/quadrille/quadgo) [![Go Report Card](https://goreportcard.com/badge/github.com/quadrille/quadgo)](https://goreportcard.com/report/github.com/quadrille/quadgo)

Quadgo is the official QuadrilleDB Go driver. It has a simple and intuitive API which enables Go developers to quickly get up and running without any hassle.

## Features

### Cluster Discovery

Quadgo offers automatic cluster discovery. Given only a single node address, it finds out cluster info and connects to the leader

### Automatic Failover Management

Quadgo will automatically failover in case a change in leader is detected

### Synchronous and Concurrent

Quadgo provides a synchronous interface and is built with high concurreny in mind. To achieve high degree of concurrency, quadgo maintains a connection pool and multiplexes multiple requests onto each of these connections. It also uses pipelining to send consecutive requests without waiting for responses to arrive for previous requests.

## Getting Started

To get the package, execute

```bash
go get github.com/quadrille/quadgo
```

To import the package, add following line to you code

```bash
import "github.com/quadrille/quadgo"
```

## Get quadgo client instance

```go
quadClient := quadgo.NewClient("localhost:5679")

```

quadgo uses Singleton to create the client. Therefore you can safely get the same client instance by calling NewClient from multiple Goroutines

### Insert Location

```go
//func (q *quadrilleClient) AddLocation(locationID string, latitude, longitude float64, data map[string]interface{}) (err error)
err := quadgo.Insert("loc123", 12.9660585, 77.7157481, map[string]interface{}{"some_key": "some_val"})
```

### Get Location

```go
//func (q *quadrilleClient) Get(locationID string) (*Location, error)
location,err := quadgo.Get("loc123")
```

### Get Nearby locations

```go
//func (q *quadrilleClient) Nearby(latitude, longitude float64, radiusInMeters, limit int) (neighbors []NeighborResult, err error)
results, err := quadgo.Nearby(12.9659327, 77.7168568, 500, 10)
```

### Delete Location

```go
//func (q *quadrilleClient) Delete(locationID string) (err error)
err := quadgo.Delete("loc123")
```

### Update Location

Complete replace

```go
//func (q *QuadrilleClient) Update(locationID string, latitude, longitude float64, data map[string]interface{}) (err error)
err := quadgo.Update("loc123", 12.9958017, 77.6942081, map[string]interface{}{"some_key": "some_val2"})
```

Update **location** only

```go
//func (q *QuadrilleClient) UpdateLocation(locationID string, latitude, longitude float64) (err error)
err := quadgo.UpdateLocation("loc123", 12.9958017, 77.6942081)
```

Update **data** only

```go
//func (q *QuadrilleClient) UpdateData(locationID string, data map[string]interface{}) (err error)
err := quadgo.UpdateData("loc123", map[string]interface{}{"some_key": "some_val3"})
```

### BulkWrite

Used to batch multiple write operations in a single QuadrilleDB operation. Currently supported operations include insert, update, updateloc, updatedata

```go
bulk := quadgo.NewBulkWrite()

bulk.Add(quadgo.NewInsertOperation("loc123", 12, 77, map[string]interface{}{}))
bulk.Add(quadgo.NewUpdateLocOperation("loc456", 17, 78)

quadClient.ExecuteBulk(bulk)
```

Use these factory methods to prepare any of the supported operations for BulkWrite

- NewInsertOperation(locationID string, lat, lon float64, data map[string]interface{})
- NewUpdateOperation(locationID string, lat, lon float64, data map[string]interface{})
- NewUpdateLocOperation(locationID string, lat, lon float64)
- NewUpdateDataOperation(locationID string, data map[string]interface{})

Each of the above factory methods returns a \*bulkWriteOperation which can be passed to a bulk.Add() call.

### Close connection

```go
quadgo.Close()
```
