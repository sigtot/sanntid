package seller

import (
	"encoding/json"
	"fmt"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/types"
	"testing"
	"time"
)

const id1 string = "firstID"
const id2 string = "secondID"

func TestSeller(t *testing.T) {
	bestPrice := 4
	betterThanBestPrice := 2
	newCalls := make(chan types.Call)
	go StartSelling(newCalls)

	bidPubChan := pubsub.StartPublisher(pubsub.BidDiscoveryPort)
	ackPubChan := pubsub.StartPublisher(pubsub.AckDiscoveryPort)
	forSaleSubChan, _ := pubsub.StartSubscriber(pubsub.SalesDiscoveryPort, pubsub.SalesTopic)
	soldToSubChan, _ := pubsub.StartSubscriber(pubsub.SoldToDiscoveryPort, pubsub.SoldToTopic)
	ackSubChan, _ := pubsub.StartSubscriber(pubsub.AckDiscoveryPort, pubsub.AckTopic)

	firstCall := types.Call{Type: types.Hall, Floor: 3, Dir: types.Down, ElevatorID: ""}
	newCalls <- firstCall

	timeOut := time.After(time.Millisecond * 200)
	for {
		select {
		case itemForSale := <-forSaleSubChan:
			item := types.Call{}
			fmt.Printf("Item for sale: %s\n", string(itemForSale[:]))
			err := json.Unmarshal(itemForSale, &item)
			if err != nil {
				panic(fmt.Sprintf("Could not unmarshal bid %s", err.Error()))
			}
			firstBid := types.Bid{Call: item, Price: 30, ElevatorID: id1}
			js, err := json.Marshal(firstBid)
			fmt.Printf("Bid: %s\n", string(js[:]))
			if err != nil {
				panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
			}
			bidPubChan <- js

			secondBid := types.Bid{Call: item, Price: bestPrice, ElevatorID: id2}
			js, err = json.Marshal(secondBid)
			fmt.Printf("Bid: %s\n", string(js[:]))
			if err != nil {
				panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
			}
			bidPubChan <- js

			otherBid := types.Bid{
				Call: types.Call{
					Type:       types.Hall,
					Floor:      4,
					Dir:        types.Down,
					ElevatorID: ""},
				Price:      betterThanBestPrice,
				ElevatorID: id2}
			js, err = json.Marshal(otherBid)
			fmt.Printf("Bid: %s\n", string(js[:]))
			if err != nil {
				panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
			}
			bidPubChan <- js
		case soldItemJson := <-soldToSubChan:
			soldItem := types.SoldTo{}
			err := json.Unmarshal(soldItemJson, &soldItem)

			if err != nil {
				panic(fmt.Sprintf("Could not unmarshal soldTo %s", err.Error()))
			}
			fmt.Printf("Sold to: %+v\n", soldItem)
			ack := types.Ack{Bid: soldItem.Bid}
			js, err := json.Marshal(ack)
			fmt.Printf("Ack: %s\n", string(js[:]))
			if err != nil {
				panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
			}
			ackPubChan <- js
		case ackJson := <-ackSubChan:
			ack := types.Ack{}
			err := json.Unmarshal(ackJson, &ack)
			if err != nil {
				panic(fmt.Sprintf("Could not unmarshal ack %s", err.Error()))
			}
			fmt.Printf("Got ack %+v\n", ack)
			if ack.Price == bestPrice {
				return
			}
			t.FailNow()
		case <-timeOut:
			t.Error("Never got an ack. Timed out")
			t.FailNow()
		}
	}
}
