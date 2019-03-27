/*
Package buyer contains the logic for bidding on and buying calls.
*/
package buyer

import (
	"encoding/json"
	"fmt"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/pubsub/subscribe"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
)

const numFloors = 4
const topFloor = numFloors - 1
const bottomFloor = 0
const moduleName = "BUYER"

type PriceCalculator interface {
	GetPrice(types.Call) int
}

// StartBuying starts a go-routine that bids on and buys calls.
// It subscribes to sale propositions and sales.
// It publishes bids and sale acknowledgements.
// A PriceCalculator interface is used to get the price on a call.
func StartBuying(priceCalc PriceCalculator, newOrders chan types.Order) {
	bidPubChan := publish.StartPublisher(pubsub.BidDiscoveryPort)
	ackPubChan := publish.StartPublisher(pubsub.AckDiscoveryPort)
	forSaleSubChan, _ := subscribe.StartSubscriber(pubsub.SalesDiscoveryPort, pubsub.SalesTopic)
	soldToSubChan, _ := subscribe.StartSubscriber(pubsub.SoldToDiscoveryPort, pubsub.SoldToTopic)

	elevatorID, _ := mac.GetMacAddr()

	var log = logrus.New()

	go func() {
		for {
			select {
			case callJson := <-forSaleSubChan:
				call := types.Call{}
				err := json.Unmarshal(callJson, &call)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal call %s", err.Error()))
				}

				if call.Type == types.Cab && call.ElevatorID != elevatorID {
					break // Do not respond to other elevator's cab calls
				}

				if call.Floor > topFloor || call.Floor < bottomFloor {
					break // Do not respond to calls outside elevator floor range
				}
				if call.Floor == topFloor && call.Dir == types.Up || call.Floor == bottomFloor && call.Dir == types.Down {
					break
				}

				price := priceCalc.GetPrice(call)
				bid := types.Bid{Call: call, Price: price, ElevatorID: elevatorID}

				js, err := json.Marshal(bid)
				if err != nil {
					panic(fmt.Sprintf("Could not marshal bid %s", err.Error()))
				}
				bidPubChan <- js

				utils.LogBid(log, moduleName, "Placed bid on order", bid)
			case soldToJson := <-soldToSubChan:
				soldTo := types.SoldTo{}
				err := json.Unmarshal(soldToJson, &soldTo)

				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal sold call %s", err.Error()))
				}

				if soldTo.ElevatorID == elevatorID {
					ack := types.Ack{Bid: soldTo.Bid}
					js, err := json.Marshal(ack)
					if err != nil {
						panic(fmt.Sprintf("Could not marshal ack %s", err.Error()))
					}
					ackPubChan <- js
					newOrders <- types.Order{Call: soldTo.Call}

					utils.LogAck(log, moduleName, "Bought order", ack)
				}
			}
		}
	}()
}
