package orderwatcher

import (
	"encoding/json"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/types"
	bolt "go.etcd.io/bbolt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const testDbName = "test.db"
const testElevID = "cb:32:f6:7e:2d:cc"
const testDbPerms = 0600
const testDbTimeout = 1000

func TestInitHallOrderBuckets(t *testing.T) {
	db, err := bolt.Open(testDbName, testDbPerms, &bolt.Options{Timeout: testDbTimeout * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
	err = initHallOrderBuckets(db)
	if err != nil {
		t.Fatal(err)
	}
	err = db.View(func(tx *bolt.Tx) error {
		upB := tx.Bucket([]byte(hallUpBucketName))
		downB := tx.Bucket([]byte(hallDownBucketName))
		if upB == nil || downB == nil {
			t.Fatal("Expected buckets to be created successfully, but they are nil after creation")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(testDbName); err != nil {
		panic(err)
	}
}

func TestStartOrderWatcher(t *testing.T) {
	db, err := bolt.Open(testDbName, testDbPerms, &bolt.Options{Timeout: testDbTimeout * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
		/*
			if err := os.Remove(testDbName); err != nil {
				panic(err)
			}
		*/
	}()

	ackPubChan := publish.StartPublisher(pubsub.AckDiscoveryPort)
	orderDelPubChan := publish.StartPublisher(pubsub.OrderDeliveredDiscoveryPort)

	callsForSale := make(chan types.Call)
	quit := make(chan int)
	var wg sync.WaitGroup
	StartOrderWatcher(callsForSale, db, quit, &wg)
	StartDbDistributor(db, testDbName, quit)

	orders := []types.Order{
		{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 1}},
		{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 6}},
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
	}

	for _, v := range orders {
		ackJson, err := json.Marshal(types.Ack{Bid: types.Bid{Call: v.Call, Price: 2, ElevatorID: testElevID}})
		if err != nil {
			t.Fatal("Could not marshal ack")
		}
		ackPubChan <- ackJson
	}

	time.Sleep(1500 * time.Millisecond)

	err = db.View(func(tx *bolt.Tx) error {
		buckets := []*bolt.Bucket{
			tx.Bucket([]byte(hallUpBucketName)),
			tx.Bucket([]byte(hallDownBucketName)),
			tx.Bucket([]byte(testElevID)),
		}
		for i, b := range buckets {
			var retrievedWT WatchThis
			if err := json.Unmarshal(b.Get([]byte(strconv.Itoa(orders[i].Floor))), &retrievedWT); err != nil {
				t.Fatal(err)
			}

			if !(retrievedWT.Call == orders[i].Call) {
				t.Fatalf("Retrieved value %+v does not match %+v\n", retrievedWT, orders[i].Call)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ordersDelivered := []types.Order{
		orders[0],
		{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 6}},
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 2, ElevatorID: "fd:34:e6:b1:33:7e"}},
	}
	for _, v := range ordersDelivered {
		orderJson, err := json.Marshal(v)
		if err != nil {
			t.Fatal("Could not marshal order")
		}
		orderDelPubChan <- orderJson
	}

	time.Sleep(1000 * time.Millisecond)
	err = db.View(func(tx *bolt.Tx) error {
		buckets := []*bolt.Bucket{
			tx.Bucket([]byte(hallUpBucketName)),
			tx.Bucket([]byte(hallDownBucketName)),
			tx.Bucket([]byte("fd:34:e6:b1:33:7e")),
		}
		for i, b := range buckets {
			if string(b.Get([]byte(strconv.Itoa(orders[i].Floor)))) != "" {
				t.Fatal("Did not get empty string in delivered order slot")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	quit <- 1
}
