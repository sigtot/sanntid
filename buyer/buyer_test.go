package buyer

import (
	"encoding/json"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/types"
	"testing"
	"time"
)

type MockPriceCalculator struct{}

func (pc *MockPriceCalculator) GetPrice(call types.Call) int {
	return 2
}

func TestBuyer(t *testing.T) {
	forSalePubChan := pubsub.StartPublisher(pubsub.SalesDiscoveryPort)
	soldToPubChan := pubsub.StartPublisher(pubsub.SoldToDiscoveryPort)

	elevatorID, err := mac.GetMacAddr()
	if err != nil {
		t.Fatalf("Could not get mac addr %s\n", err.Error())
	}

	priceCalc := MockPriceCalculator{}
	newOrders := make(chan types.Order)
	StartBuying(&priceCalc, newOrders)

	// Sell call
	call := types.Call{Type: types.Hall, Floor: 3, Dir: types.Down, ElevatorID: ""}
	callJson, err := json.Marshal(call)
	if err != nil {
		t.Fatalf("Could not marshal call %s\n", err.Error())
	}
	forSalePubChan <- callJson

	time.Sleep(20 * time.Millisecond)

	// Send sold to
	soldTo := types.SoldTo{Bid: types.Bid{
		Call:       call,
		Price:      priceCalc.GetPrice(call),
		ElevatorID: elevatorID,
	}}
	soldToJson, err := json.Marshal(soldTo)
	if err != nil {
		t.Fatalf("Could not marshal soldTo %s\n", err.Error())
	}
	soldToPubChan <- soldToJson

	time.Sleep(20 * time.Millisecond)
	select {
	case newOrder := <-newOrders:
		order := types.Order{Call: call}
		if newOrder != order {
			t.Fatalf("Bad order received: %+v\n", order)
		}
	case <-time.After(20 * time.Millisecond):
		t.Fatal("Timed out waiting for bought order")
	}
}
