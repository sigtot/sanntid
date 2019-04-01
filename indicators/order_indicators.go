/*
Package indicators contains procedures for listening for events on the network and passively updating the
order indicators accordingly.
*/
package indicators

import (
	"encoding/json"
	"fmt"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	"sync"
)

const numFloors = 4 // TODO: Move this maybe
const topFloor = numFloors - 1
const bottomFloor = 0
const moduleName = "ORDER IND"

// StartIndicatorHandler starts a go-routine that initializes the indicators, and listens for call sales and
// order deliveries on the network, updating the order indicators accordingly.
// An indicator handler subscribes to sale acknowledgements and order deliveries.
func StartIndicatorHandler(quit <-chan int, wg *sync.WaitGroup) {
	ackSubChan, _ := pubsub.StartSubscriber(pubsub.AckDiscoveryPort, pubsub.AckTopic)
	orderDeliveredSubChan, _ := pubsub.StartSubscriber(pubsub.OrderDeliveredDiscoveryPort, pubsub.OrderDeliveredTopic)
	allOff()
	macAddr, err := mac.GetMacAddr()
	if err != nil {
		panic(err)
	}
	log := logrus.New()
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case ackJson := <-ackSubChan:
				ack := types.Ack{}
				err := json.Unmarshal(ackJson, &ack)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal ack %s", err.Error()))
				}

				withinRange := ack.Call.Floor <= topFloor || ack.Call.Floor >= bottomFloor
				if withinRange && (ack.Call.Type == types.Hall || ack.ElevatorID == macAddr) {
					elevio.SetButtonLamp(getBtnType(ack.Call.Type, ack.Call.Dir), ack.Call.Floor, true)
				}

			case orderJson := <-orderDeliveredSubChan:
				utils.Log(log, moduleName, "Got order delivered")
				order := types.Order{}
				err := json.Unmarshal(orderJson, &order)
				if err != nil {
					panic(fmt.Sprintf("Could not unmarshal order %s", err.Error()))
				}
				withinRange := order.Floor <= topFloor || order.Floor >= bottomFloor
				if withinRange && (order.Type == types.Hall || order.ElevatorID == macAddr) {
					elevio.SetButtonLamp(getBtnType(order.Type, order.Dir), order.Floor, false)
				}
			case <-quit:
				allOff()
				utils.Log(log, moduleName, "Turned off all order indicators")
				return
			}
		}
	}()
}

// getBtnType translates a call type and a call direction to an elevio button type.
func getBtnType(callType types.CallType, dir types.Direction) elevio.ButtonType {
	if callType == types.Hall {
		if dir == types.Up {
			return elevio.BtnHallUp
		}
		return elevio.BtnHallDown
	}
	return elevio.BtnCab
}

// allOff turns off all order indicators.
func allOff() {
	for i := bottomFloor; i <= topFloor; i++ {
		elevio.SetButtonLamp(elevio.BtnCab, i, false)
		if i != bottomFloor {
			elevio.SetButtonLamp(elevio.BtnHallDown, i, false)
		}
		if i != topFloor {
			elevio.SetButtonLamp(elevio.BtnHallUp, i, false)
		}
	}
}
