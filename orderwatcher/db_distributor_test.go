package orderwatcher

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/sigtot/sanntid/pubsub"
	bolt "go.etcd.io/bbolt"
	"io"
	"os"
	"testing"
	"time"
)

func TestStartDbDistributor(t *testing.T) {
	db, err := bolt.Open(testDbName, testDbPerms, &bolt.Options{Timeout: testDbTimeout * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	if err := writeToDb(db, hallUpBucketName, "1", []byte("skkrt")); err != nil {
		t.Fatal("Could not write to db")
	}

	dbSubChan, _ := pubsub.StartSubscriber(pubsub.DbDiscoveryPort, pubsub.DbDiscoveryTopic)

	quit := make(chan int)
	StartDbDistributor(db, testDbName, quit)

	dbMsgJson := <-dbSubChan
	dbMsg := dbMsg{}
	err = json.Unmarshal(dbMsgJson, &dbMsg)
	if err != nil {
		panic(fmt.Sprintf("Could not unmarshal db message %s", err.Error()))
	}

	fmt.Printf("%+v\n", dbMsg)

	var buf bytes.Buffer
	buf.Write(dbMsg.Buf)
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("testDB.db")
	if _, err = io.Copy(f, zr); err != nil {
		panic(err)
	}

	db2, err := bolt.Open("testDB.db", testDbPerms, &bolt.Options{Timeout: testDbTimeout * time.Millisecond})
	if err != nil {
		panic(err)
	}

	err = db2.Close()
	if err != nil {
		panic(err)
	}

	// TODO: What's this?
	for {
		select {
		case <-dbSubChan:
			println("Got state")
		}
	}
}
