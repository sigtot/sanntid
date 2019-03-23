package seller

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/sigtot/sanntid/hotchan"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/pubsub/subscribe"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"time"
)

const TTL = 400

// Seller states
const (
	Idle = iota
	WaitingForBids
	WaitingForAck
)

const BiddingRoundDuration = time.Millisecond * 10
const AckWaitDuration = time.Millisecond * 10

// StartSelling starts the seller.
// It will attempt to sell calls sent into the newCalls channel on the network.
func StartSelling(newCalls chan types.Call) {
	state := Idle

	forSalePubChan := publish.StartPublisher(pubsub.SalesDiscoveryPort)
	soldToPubChan := publish.StartPublisher(pubsub.SoldToDiscoveryPort)
	bidSubChan, _ := subscribe.StartSubscriber(pubsub.BidDiscoveryPort)
	ackSubChan, _ := subscribe.StartSubscriber(pubsub.AckDiscoveryPort)

	var log = logrus.New()

	forSale := hotchan.HotChan{}
	forSale.Start()

	defer forSale.Stop()

	go func() {
		for {
			val := <-newCalls
			hcItem := hotchan.Item{Val: val, TTL: TTL * time.Millisecond}
			forSale.Insert(hcItem)
		}
	}()
	var itemForSale hotchan.Item
	var lowestBid types.Bid
	for {
		switch state {
		case Idle:
			for {
				itemForSale = <-forSale.Out
				// Announce sale on network
				js, err := json.Marshal(itemForSale.Val)
				if err != nil {
					panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
				}
				forSalePubChan <- js

				utils.LogCall(log, "SELLER", "Started a new sale", itemForSale.Val.(types.Call))
				state = WaitingForBids
				break
			}
		case WaitingForBids:
			var recvBids []types.Bid
			timeOut := time.After(BiddingRoundDuration)
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

					utils.LogBid(log, "SELLER", "Received bid", bid)
				case <-timeOut:
					if len(recvBids) == 0 {
						//Try to sell again
						forSale.Insert(itemForSale)
						state = Idle
						break L1
					}
					lowestBid = getLowestBid(recvBids)

					js, err := json.Marshal(lowestBid)
					if err != nil {
						panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
					}
					soldToPubChan <- js
					state = WaitingForAck
					break L1
				}
			}
		case WaitingForAck:
			timeOut := time.After(AckWaitDuration)
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
						utils.LogAck(log, "SELLER", "Got ack from lowest bidder", ack)
						state = Idle
						break L2
					}
				case <-timeOut:
					forSale.Insert(itemForSale)
					state = Idle
					break L2
				}
			}
		}
	}
}

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
