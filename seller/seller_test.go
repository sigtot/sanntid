package seller

import (
	"encoding/json"
	"fmt"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/pubsub/subscribe"
	"github.com/sigtot/sanntid/types"
	"testing"
	"time"
)

const id1 string = "firstID"
const id2 string = "secondID"

func TestSeller(t *testing.T) {
	bestPrice := 4
	newCalls := make(chan types.Call)
	go StartSelling(newCalls)

	bidPublishChan := make(chan []byte, 1024)
	go publish.StartPublisher(pubsub.BidDiscoveryPort, bidPublishChan)

	ackPublishChan := make(chan []byte)
	go publish.StartPublisher(pubsub.AckDiscoveryPort, ackPublishChan)
	<-time.After(30 * time.Millisecond)
	forSaleSubChan, _ := subscribe.StartSubscriber(pubsub.SalesDiscoveryPort)

	soldToSubChan, _ := subscribe.StartSubscriber(pubsub.SoldToDiscoveryPort)

	ackSubChan, _ := subscribe.StartSubscriber(pubsub.AckDiscoveryPort)

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
			bidPublishChan <- js

			secondBid := types.Bid{Call: item, Price: bestPrice, ElevatorID: id2}
			js, err = json.Marshal(secondBid)
			fmt.Printf("Bid: %s\n", string(js[:]))
			if err != nil {
				panic(fmt.Sprintf("Could not marshal call %s", err.Error()))
			}
			bidPublishChan <- js
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
			ackPublishChan <- js
		case ackJson := <-ackSubChan:
			ack := types.Ack{}
			err := json.Unmarshal(ackJson, &ack)
			if err != nil {
				panic(fmt.Sprintf("Could not unmarshal ack %s", err.Error()))
			}
			fmt.Printf("Got ack %+v\n", ack)
			if ack.Price == bestPrice {
				return
			} else {
				t.FailNow()
			}
		case <-timeOut:
			t.Error("Never got an ack. Timed out")
			t.FailNow()
		}
	}
}
