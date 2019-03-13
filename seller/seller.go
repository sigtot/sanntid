package seller

import (
	"encoding/json"
	"fmt"
	"github.com/sigtot/sanntid/hotchan"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/pubsub/subscribe"
	"github.com/sigtot/sanntid/types"
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
	State := Idle

	// Start sales publisher
	forSalePubChan := make(chan []byte)
	publish.StartPublisher(pubsub.SalesDiscoveryPort, forSalePubChan)

	// Start sold to publisher
	soldToPubChan := make(chan []byte)
	publish.StartPublisher(pubsub.SoldToDiscoveryPort, soldToPubChan)

	// Start bid subscriber
	bidSubChan, _ := subscribe.StartSubscriber(pubsub.BidDiscoveryPort)

	// Start ack subscriber
	ackSubChan, _ := subscribe.StartSubscriber(pubsub.AckDiscoveryPort)

	forSale := hotchan.HotChan{}
	forSale.Start()

	defer forSale.Stop()

	go func() {
		for {
			val := <-newCalls
			hcItem := hotchan.Item{Val: val, TTL: TTL * time.Millisecond}
			forSale.In <- hcItem
		}
	}()
	var itemForSale hotchan.Item
	var lowestBid types.Bid
	for {
		switch State {
		case Idle:
			for {
				itemForSale = <-forSale.Out
				// Announce sale on network
				js, err := json.Marshal(itemForSale.Val)
				if err != nil {
					panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
				}
				forSalePubChan <- js
				State = WaitingForBids
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
					err := json.Unmarshal(bidJson, bid)
					if err != nil {
						panic(fmt.Sprintf("Could not unmarshal bid %s", err.Error()))
					}
					if bid.Call == itemForSale.Val {
						recvBids = append(recvBids, bid)
					}
				case <-timeOut:
					if len(recvBids) == 0 {
						//Try to sell again
						forSale.In <- itemForSale
					}
					lowestBid = getLowestBid(recvBids)

					js, err := json.Marshal(lowestBid)
					if err != nil {
						panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
					}
					soldToPubChan <- js
					State = WaitingForAck
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
					err := json.Unmarshal(ackJson, ack)
					if err != nil {
						panic(fmt.Sprintf("Could not unmarshal ack %s", err.Error()))
					}
					if ack.Bid == lowestBid {
						break L2
					}
				case <-timeOut:
					forSale.In <- itemForSale
					break L2
				}
			}
		}
	}
}

func getLowestBid(bids []types.Bid) types.Bid {
	var lowestBid types.Bid
	lowestPrice := bids[0].Price
	for _, bid := range bids {
		if bid.Price < lowestPrice {
			lowestPrice = bid.Price
			lowestBid = bid
		}
	}
	return lowestBid
}
