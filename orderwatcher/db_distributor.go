package orderwatcher

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	bolt "go.etcd.io/bbolt"
	"io"
	"os"
	"time"
)

const dbDistributeInterval = 10000

type DbMsg struct {
	Buf        []byte
	ElevatorID string
}

func StartDbDistributor(db *bolt.DB, dbName string, quit <-chan int) chan int {
	dbPubChan := publish.StartPublisher(pubsub.DbDiscoveryPort)
	elevatorID, _ := mac.GetMacAddr()
	quitAck := make(chan int)

	go func() {
		dbDistributeTicker := time.NewTicker(dbDistributeInterval * time.Millisecond)
		for {
			select {
			case <-dbDistributeTicker.C:
				buf, err := copyDb(db, dbName)
				if err != nil {
					panic(err)
				}

				dbMsg := DbMsg{Buf: buf.Bytes(), ElevatorID: elevatorID}
				dbJson, err := json.Marshal(dbMsg)
				if err != nil {
					panic("Could not marshal buffer")
				}

				dbPubChan <- dbJson
			case <-quit:
				quitAck <- 0
			}
		}
	}()
	return quitAck
}

func copyDb(db *bolt.DB, dbName string) (buf bytes.Buffer, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		f, err := os.Open(dbName)
		if err != nil {
			return err
		}
		defer func() { err = f.Close() }()

		zw := gzip.NewWriter(&buf)
		defer func() { err = zw.Close() }()

		if _, err = io.Copy(zw, f); err != nil {
			return err
		}
		return nil
	})
	return buf, err
}
