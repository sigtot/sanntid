package indicators

import (
	"encoding/json"
	"fmt"
	"github.com/TTK4145/driver-go/elevio"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/types"
	"log"
	"testing"
)

func TestStartHandlingIndicators(t *testing.T) {
	elevio.Init("localhost:15657", 4)
	StartHandlingIndicators()
	ackPubChan := publish.StartPublisher(pubsub.AckDiscoveryPort)
	orderDeliveredPubChan := publish.StartPublisher(pubsub.OrderDeliveredDiscoveryPort)
	call := types.Call{Type: types.Cab, Floor: 4, Dir: types.InvalidDir, ElevatorID: ""}
	order1 := types.Order{Call: call}
	bid1 := types.Bid{Call: call, Price: 1, ElevatorID: ""}
	ack1 := types.Ack{Bid: bid1}
	js, err := json.Marshal(ack1)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Could not marshal ack %s", err.Error()))
	}
	ackPubChan <- js

	js, err = json.Marshal(order1)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Could not marshal order %s", err.Error()))
	}
	orderDeliveredPubChan <- js
}
