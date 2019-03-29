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

type dbMsg struct {
	buf      []byte
	senderID string
}

// StartDbDistributor starts distributing the database of orders.
// It compresses the file and publishes it as a DbMsg on the network.
func StartDbDistributor(db *bolt.DB, dbName string, quit <-chan int) chan int {
	dbPubChan := publish.StartPublisher(pubsub.DbDiscoveryPort)
	elevatorID, _ := mac.GetMacAddr()
	quitAck := make(chan int)

	go func() {
		dbDistributeTicker := time.NewTicker(dbDistributeInterval * time.Millisecond)
		for {
			select {
			case <-dbDistributeTicker.C:
				buf, err := getCompressesCopyDb(db, dbName)
				if err != nil {
					panic(err)
				}

				dbMsg := dbMsg{buf: buf.Bytes(), senderID: elevatorID}
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

// getCompressesCopyDb returns a compressed buffer copy of the input database.
func getCompressesCopyDb(db *bolt.DB, dbName string) (buf bytes.Buffer, err error) {
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
