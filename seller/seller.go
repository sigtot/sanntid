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
	state := Idle

	// Start sales publisher
	forSalePubChan := publish.StartPublisher(pubsub.SalesDiscoveryPort)

	// Start sold to publisher
	soldToPubChan := publish.StartPublisher(pubsub.SoldToDiscoveryPort)
	<-time.After(30 * time.Millisecond)
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
					fmt.Println("Got bid")
					if bid.Call == itemForSale.Val {
						recvBids = append(recvBids, bid)
					}
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
