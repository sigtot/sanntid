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

const moduleName = "BUYER"

type PriceCalculator interface {
	GetPrice(types.Call) int
}

func StartBuying(priceCalc PriceCalculator, newOrders chan types.Order) {
	bidPubChan := publish.StartPublisher(pubsub.BidDiscoveryPort)
	ackPubChan := publish.StartPublisher(pubsub.AckDiscoveryPort)
	forSaleSubChan, _ := subscribe.StartSubscriber(pubsub.SalesDiscoveryPort, pubsub.SalesTopic)
	soldToSubChan, _ := subscribe.StartSubscriber(pubsub.SoldToDiscoveryPort, pubsub.SalesTopic)

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
