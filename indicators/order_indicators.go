package indicators

import (
	"encoding/json"
	"fmt"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/subscribe"
	"github.com/sigtot/sanntid/types"
)

func StartHandlingIndicators() {
	ackSubChan, _ := subscribe.StartSubscriber(pubsub.AckDiscoveryPort)
	orderDeliveredSubChan, _ := subscribe.StartSubscriber(pubsub.OrderDeliveredDiscoveryPort)
	go func() {
		for {
			select {
			case ackJson := <-ackSubChan:
				ack := types.Ack{}
				err := json.Unmarshal(ackJson, &ack)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal ack %s", err.Error()))
				}
				elevio.SetButtonLamp(getBtnType(ack.Call.Type, ack.Call.Dir), ack.Call.Floor, true)

			case orderJson := <-orderDeliveredSubChan:
				println("Entered orderJson")
				order := types.Order{}
				err := json.Unmarshal(orderJson, &order)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal order %s", err.Error()))
				}
				elevio.SetButtonLamp(getBtnType(order.Type, order.Dir), order.Floor, false)
			}
		}
	}()
}

func getBtnType(callType types.CallType, dir types.Direction) elevio.ButtonType {
	if callType == types.Hall {
		if dir == types.Up {
			return elevio.BtnHallUp
		} else {
			return elevio.BtnHallDown
		}
	} else {
		return elevio.BtnCab
	}
}
