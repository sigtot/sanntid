package orderwatcher

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/utils"
	bolt "go.etcd.io/bbolt"
	"io"
	"os"
	"time"
)

const dbDistributeInterval = 10000

type dbMsg struct {
	Buf      []byte
	SenderID string
}

// StartDbDistributor starts distributing the database of orders.
// It compresses the file and publishes it as a DbMsg on the network.
func StartDbDistributor(db *bolt.DB, dbName string, quit <-chan int) chan int {
	dbPubChan := pubsub.StartPublisher(pubsub.DbDiscoveryPort)
	elevatorID, err := mac.GetMacAddr()
	utils.OkOrPanic(err)
	quitAck := make(chan int)

	go func() {
		dbDistributeTicker := time.NewTicker(dbDistributeInterval * time.Millisecond)
		for {
			select {
			case <-dbDistributeTicker.C:
				buf, err := getCompressesCopyDb(db, dbName)
				utils.OkOrPanic(err)

				dbMsg := dbMsg{Buf: buf.Bytes(), SenderID: elevatorID}
				dbJson, err := json.Marshal(dbMsg)
				utils.OkOrPanic(err)

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
