/*
Package seller contains logic for announcing sale propositions,
running bidding rounds and selling to the lowest bidder.
*/
package seller

import (
	"encoding/json"
	"fmt"
	"github.com/sigtot/sanntid/hotchan"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/pubsub/subscribe"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	"time"
)

const ttl = 400

// Seller states
const (
	idle = iota
	waitingForBids
	waitingForAck
)

const biddingRoundDuration = time.Millisecond * 10
const ackWaitDuration = time.Millisecond * 10
const moduleName = "SELLER"

// StartSelling starts a seller that sells calls, runs bidding rounds and sells to the lowest bidder.
// A seller subscribes to bids and sale acknowledgements.
// A seller publishes sale propositions and sales.
func StartSelling(newCalls chan types.Call) {
	state := idle

	forSalePubChan := publish.StartPublisher(pubsub.SalesDiscoveryPort)
	soldToPubChan := publish.StartPublisher(pubsub.SoldToDiscoveryPort)
	bidSubChan, _ := subscribe.StartSubscriber(pubsub.BidDiscoveryPort, pubsub.BidTopic)
	ackSubChan, _ := subscribe.StartSubscriber(pubsub.AckDiscoveryPort, pubsub.AckTopic)

	var log = logrus.New()

	forSale := hotchan.HotChan{}
	forSale.Start()

	defer forSale.Stop()

	go func() {
		for {
			val := <-newCalls
			hcItem := hotchan.Item{Val: val, TTL: ttl * time.Millisecond}
			forSale.Insert(hcItem)
		}
	}()
	var itemForSale hotchan.Item
	var lowestBid types.Bid
	for {
		switch state {
		case idle:
			for {
				itemForSale = <-forSale.Out
				// Announce sale on network
				js, err := json.Marshal(itemForSale.Val)
				if err != nil {
					panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
				}
				forSalePubChan <- js

				utils.LogCall(log, moduleName, "Started a new sale", itemForSale.Val.(types.Call))
				state = waitingForBids
				break
			}
		case waitingForBids:
			var recvBids []types.Bid
			timeOut := time.After(biddingRoundDuration)
		L1:
			for {
				select {
				case bidJson := <-bidSubChan:
					bid := types.Bid{}
					err := json.Unmarshal(bidJson, &bid)
					if err != nil {
						panic(fmt.Sprintf("Could not unmarshal bid %s", err.Error()))
					}
					if bid.Call == itemForSale.Val {
						recvBids = append(recvBids, bid)
					}

					utils.LogBid(log, moduleName, "Received bid", bid)
				case <-timeOut:
					if len(recvBids) == 0 {
						//Try to sell again
						forSale.Insert(itemForSale)
						state = idle
						break L1
					}
					lowestBid = getLowestBid(recvBids)

					js, err := json.Marshal(lowestBid)
					if err != nil {
						panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
					}
					soldToPubChan <- js
					state = waitingForAck
					break L1
				}
			}
		case waitingForAck:
			timeOut := time.After(ackWaitDuration)
		L2:
			for {
				select {
				case ackJson := <-ackSubChan:
					ack := types.Ack{}
					err := json.Unmarshal(ackJson, &ack)
					if err != nil {
						panic(fmt.Sprintf("Could not unmarshal ack %s", err.Error()))
					}
					if ack.Bid == lowestBid {
						utils.LogAck(log, moduleName, "Got ack from lowest bidder", ack)
						state = idle
						break L2
					}
				case <-timeOut:
					forSale.Insert(itemForSale)
					state = idle
					break L2
				}
			}
		}
	}
}

// GetLowestBid returns the lowest bid from a slice of bids.
func getLowestBid(bids []types.Bid) types.Bid {
	lowestBid := bids[0]
	lowestPrice := bids[0].Price
	for _, bid := range bids {
		if bid.Price < lowestPrice {
			lowestPrice = bid.Price
			lowestBid = bid
		}
	}
	return lowestBid
}
