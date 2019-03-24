package orderwatcher

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	bolt "github.com/etcd-io/bbolt"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/subscribe"
	"github.com/sigtot/sanntid/types"
	"math/rand"
	"strconv"
	"time"
)

const hallUpBucketName = "hall_up"
const hallDownBucketName = "hall_down"

const dbTraversalInterval = 500

const baseTTD = 10000
const randTTDOffset = 2000

type WatchThis struct {
	ElevatorID string
	Time       time.Time
	Call       types.Call
}

func StartOrderWatcher(callsForSale chan types.Call, db *bolt.DB) {
	okOrPanic(initHallOrderBuckets(db))

	ackSubChan, _ := subscribe.StartSubscriber(pubsub.AckDiscoveryPort)
	orderDeliveredSubChan, _ := subscribe.StartSubscriber(pubsub.OrderDeliveredDiscoveryPort)

	log := logrus.New()
	go func() {
		dbTraversalTicker := time.NewTicker(dbTraversalInterval * time.Millisecond)
		for {
			select {
			case ackJson := <-ackSubChan:
				// Unmarshal ack
				ack := types.Ack{}
				err := json.Unmarshal(ackJson, &ack)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal ack %s", err.Error()))
				}
				wt := WatchThis{ElevatorID: ack.ElevatorID, Time: time.Now(), Call: ack.Call}
				wtJson, err := json.Marshal(wt)
				if err != nil {
					panic("Could not marshal WatchThis")
				}

				// Save wt in db
				var bName string
				if wt.Call.Type == types.Hall {
					if wt.Call.Dir == types.Up {
						bName = hallUpBucketName
					} else if wt.Call.Dir == types.Down {
						bName = hallDownBucketName
					} else {
						panic("unexpected non-direction")
					}
				} else if wt.Call.Type == types.Cab {
					bName = wt.ElevatorID
					err = db.Update(func(tx *bolt.Tx) error {
						_, err := tx.CreateBucketIfNotExists([]byte(bName))
						return err
					})
					okOrPanic(err)
				}
				err = db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(bName))

					err = b.Put([]byte(strconv.Itoa(wt.Call.Floor)), []byte(wtJson))
					return err
				})
				okOrPanic(err)

			case orderJson := <-orderDeliveredSubChan:
				// TODO: Implement
				order := types.Order{}
				err := json.Unmarshal(orderJson, &order)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal order %s", err.Error()))
				}
			case <-dbTraversalTicker.C:
				err := db.View(func(tx *bolt.Tx) error {
					return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
						err := b.ForEach(func(k []byte, v []byte) error {
							if wt, err := unmarshalWatchThis(v); err == nil {
								if time.Now().After(wt.Time.Add(getTTD())) {
									// Resell order
									callsForSale <- wt.Call

									// Update time
									wt.Time = time.Now()
									if wtJson, err := json.Marshal(wt); err == nil {
										return b.Put(k, wtJson)
									}

									logWatchThis(log, "ORDER WATCHER", "Sent order to seller for resale", *wt)
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
				okOrPanic(err)
			}
		}
	}()
}

func unmarshalWatchThis(wtJson []byte) (*WatchThis, error) {
	wt := WatchThis{}
	err := json.Unmarshal(wtJson, &wt)
	return &wt, err
}

func initHallOrderBuckets(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(hallUpBucketName)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(hallDownBucketName)); err != nil {
			return err
		}
		return nil
	})
}

func okOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// Returns time to delivery for order, randomly distributed around its base time
func getTTD() time.Duration {
	return time.Duration(baseTTD+(rand.Intn(randTTDOffset)-randTTDOffset/2)) * time.Millisecond
}

func logWatchThis(log *logrus.Logger, moduleName string, info string, wt WatchThis) {
	log.WithFields(logrus.Fields{
		"type":  wt.Call.Type,
		"floor": wt.Call.Floor,
		"dir":   wt.Call.Dir,
		"id":    wt.ElevatorID,
		"time":  wt.Time,
	}).Infof("%-15s %s", moduleName, info)
}
