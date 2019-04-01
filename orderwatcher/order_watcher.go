/*
Package orderwatcher contains the watchdog and state distribution system, and thus handles most of the fault tolerance
in the elevator system. Orders on the network are stored in a local database,
which is regularly distributed and synced with the other elevators on the network.
*/
package orderwatcher

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

const hallUpBucketName = "hall_up"
const hallDownBucketName = "hall_down"

const dbTraversalInterval = 500

const baseTTD = 10000
const randTTDOffset = 2000

const moduleName = "ORDER WATCHER"
const logString = "%-15s%s"

const dbCopyDir = "/tmp"
const dbCopyName = "orderwatcher_copy.db"
const dbCopyPerms = 0600
const dbCopyTimeout = 500

type assignedOrder struct {
	OwnerID    string
	AssignTime time.Time
	Call       types.Call
}

// StartOrderWatcher starts the order watcher, which listens for call sales and order deliveries on the network,
// and updates a local database that stores all orders.
// It traverses the database at regular intervals and sends orders that take too long to deliver to the seller.
// The order watcher also listens for database files sent by the other db distributors
// and synchronizes them with the local database.
// An order watcher subscribes to sale acknowledgements, order deliveries and db distribution messages.
func StartOrderWatcher(callsForSale chan types.Call, db *bolt.DB, quit <-chan int, wg *sync.WaitGroup) {
	ackSubChan, _ := pubsub.StartSubscriber(pubsub.AckDiscoveryPort, pubsub.AckTopic)
	orderDeliveredSubChan, _ := pubsub.StartSubscriber(pubsub.OrderDeliveredDiscoveryPort, pubsub.OrderDeliveredTopic)
	dbSubChan, _ := pubsub.StartSubscriber(pubsub.DbDiscoveryPort, pubsub.DbDiscoveryTopic)

	elevatorID, _ := mac.GetMacAddr()
	log := logrus.New()
	wg.Add(1)
	go func() {
		defer wg.Done()

		dbTraversalTicker := time.NewTicker(dbTraversalInterval * time.Millisecond)
		defer dbTraversalTicker.Stop()
		for {
			select {
			case ackJson := <-ackSubChan:
				// Unmarshal ack and translate to assignedOrder
				ack := types.Ack{}
				err := json.Unmarshal(ackJson, &ack)
				utils.OkOrPanic(err)
				ao := assignedOrder{OwnerID: ack.ElevatorID, AssignTime: time.Now(), Call: ack.Call}
				aoJson, err := json.Marshal(ao)
				utils.OkOrPanic(err)

				// Save assignedOrder in db
				bName, err := getBucketName(ao)
				utils.OkOrPanic(err)
				err = writeToDb(db, bName, strconv.Itoa(ao.Call.Floor), aoJson)
				utils.OkOrPanic(err)

			case orderJson := <-orderDeliveredSubChan:
				// Unmarshal order
				order := types.Order{}
				err := json.Unmarshal(orderJson, &order)
				utils.OkOrPanic(err)

				// Remove order from database
				ao := assignedOrder{OwnerID: order.ElevatorID, AssignTime: time.Now(), Call: order.Call}
				bName, err := getBucketName(ao)
				utils.OkOrPanic(err)
				err = writeToDb(db, bName, strconv.Itoa(ao.Call.Floor), []byte{})
				utils.OkOrPanic(err)
			case <-dbTraversalTicker.C:
				// Traverse database and identify orders not delivered in time
				err := db.Update(func(tx *bolt.Tx) error {
					return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
						err := b.ForEach(func(k []byte, v []byte) error {
							if string(v) == "" {
								return nil
							}
							if ao, err := unmarshalAssignedOrder(v); err == nil {
								if time.Now().After(ao.AssignTime.Add(getTTD())) {
									// Resell order
									callsForSale <- ao.Call
									logAssignedOrder(log, moduleName, "Sent order to seller for resale", *ao)

									// Update time
									ao.AssignTime = time.Now()
									if aoJson, err := json.Marshal(ao); err == nil {
										return b.Put(k, aoJson)
									}

									return err
								}
							} else {
								return err
							}
							return nil
						})
						return err
					})
				})
				utils.OkOrPanic(err)
			case dbMsgJson := <-dbSubChan:
				// Unmarshal db message
				dbMsg := dbMsg{}
				err := json.Unmarshal(dbMsgJson, &dbMsg)
				utils.OkOrPanic(err)
				if dbMsg.SenderID == elevatorID {
					break // No need to sync with local db
				}
				timeBefore := time.Now()
				// Uncompress db file
				var buf bytes.Buffer
				buf.Write(dbMsg.Buf)
				zr, err := gzip.NewReader(&buf)
				utils.OkOrPanic(err)

				// Copy received db file
				f, err := os.Create(fmt.Sprintf("%s/%s", dbCopyDir, dbCopyName))
				utils.OkOrPanic(err)
				if _, err = io.Copy(f, zr); err != nil {
					panic(err)
				}
				err = zr.Close()
				utils.OkOrPanic(err)
				err = f.Close()
				utils.OkOrPanic(err)

				// Open db from copied db file
				dbCopy, err := bolt.Open(fmt.Sprintf("%s/%s", dbCopyDir, dbCopyName), dbCopyPerms, &bolt.Options{Timeout: dbCopyTimeout * time.Millisecond})
				utils.OkOrPanic(err)

				// Do union of received db and local db to sync state
				err = db.Update(func(tx *bolt.Tx) error {
					return dbCopy.View(func(txCopy *bolt.Tx) error {
						return txCopy.ForEach(func(name []byte, bCopy *bolt.Bucket) error {
							if _, err := tx.CreateBucketIfNotExists(name); err != nil {
								return err
							}

							b := tx.Bucket(name)
							return bCopy.ForEach(func(k []byte, v []byte) error {
								if string(v) == "" {
									return nil
								}
								err := b.Put(k, v)
								return err
							})
						})
					})
				})
				utils.OkOrPanic(err)
				err = dbCopy.Close()
				utils.OkOrPanic(err)
				computationDuration := time.Now().Sub(timeBefore)
				log.WithFields(logrus.Fields{
					"took": fmt.Sprintf("%.3fs", computationDuration.Seconds()),
				}).Infof(logString, moduleName, "Received db and synced")

			case <-quit:
				utils.Log(log, moduleName, "And now my watch is ended")
				return
			}
		}
	}()
}

func writeToDb(db *bolt.DB, bName string, key string, value []byte) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bName))
		return err
	}); err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bName))
		return b.Put([]byte(key), []byte(value))
	})
}

func getBucketName(ao assignedOrder) (name string, err error) {
	if ao.Call.Type == types.Hall {
		if ao.Call.Dir == types.Up {
			name = hallUpBucketName
		} else if ao.Call.Dir == types.Down {
			name = hallDownBucketName
		} else {
			err = errors.New("unexpected non-direction")
		}
	} else if ao.Call.Type == types.Cab {
		name = ao.OwnerID
	} else {
		err = errors.New("call of unexpected type")
	}
	return name, err
}
func unmarshalAssignedOrder(aoJson []byte) (*assignedOrder, error) {
	ao := assignedOrder{}
	err := json.Unmarshal(aoJson, &ao)
	return &ao, err
}

// Returns time to delivery for order, randomly distributed around its base time
func getTTD() time.Duration {
	return time.Duration(baseTTD+(rand.Intn(randTTDOffset)-randTTDOffset/2)) * time.Millisecond
}

func logAssignedOrder(log *logrus.Logger, moduleName string, info string, ao assignedOrder) {
	log.WithFields(logrus.Fields{
		"type":  ao.Call.Type,
		"floor": ao.Call.Floor,
		"dir":   ao.Call.Dir,
		"id":    ao.OwnerID,
		"time":  ao.AssignTime,
	}).Infof(logString, moduleName, info)
}
