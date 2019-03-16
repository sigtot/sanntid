package indicators

import (
	"encoding/json"
	"fmt"
	"github.com/TTK4145/driver-go/elevio"
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

				if ack.Call.Type == types.Hall {
					if ack.Call.Dir == types.Up {
						elevio.SetButtonLamp(elevio.BT_HallUp, ack.Call.Floor, true)
					} else {
						elevio.SetButtonLamp(elevio.BT_HallDown, ack.Call.Floor, true)
					}
				} else {
					elevio.SetButtonLamp(elevio.BT_Cab, ack.Call.Floor, true)
				}
			case orderJson := <-orderDeliveredSubChan:
				println("Entered orderJson")
				order := types.Order{}
				err := json.Unmarshal(orderJson, &order)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal order %s", err.Error()))
				}
				if order.Type == types.Hall {
					if order.Dir == types.Up {
						elevio.SetButtonLamp(elevio.BT_HallUp, order.Floor, false)
					} else {
						elevio.SetButtonLamp(elevio.BT_HallDown, order.Floor, false)
					}
				} else {
					elevio.SetButtonLamp(elevio.BT_Cab, order.Floor, false)
				}
			}
		}
	}()
}
