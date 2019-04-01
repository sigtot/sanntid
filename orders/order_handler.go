/*
Package orders contains functions for sorting lists of orders, calculating prices of new orders and
procedures for keeping track of the order queue and handling goal floor arrivals and new orders.
*/
package orders

import (
	"encoding/json"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	"time"
)

const numFloors = 4
const topFloor = numFloors - 1
const bottomFloor = 0

const counterDelay = 12000
const tickInterval = 1000
const deliveryDelayWeight = 1

const moduleName = "ORDER HANDLER"

// OrderHandler contains a unsorted list of all orders this elevator has to deliver. It also has a interface to the
// elevator controller in order to receive the direction and position of the elevator. Finally it has a
// delayed counter, which is added to the price calculation to penalize late deliveries. This makes the system
// robust against motor failure and similar.
type OrderHandler struct {
	orders         []types.Order
	delayedCounter utils.DelayedCounter
	elev           ElevInterface
}

// ElevInterface is used by the order handler to get the current position and direction of the elevator.
type ElevInterface interface {
	GetDir() elevio.MotorDirection
	GetPos() float64
}

// StartOrderHandler start a go-routine that sends the next goal floor on the currentGoals channel,
// when new orders are received or the elevator arrives at the current goal floor.
func StartOrderHandler(
	currentGoals chan types.Order,
	arrivals chan types.Order,
	elev ElevInterface) (*OrderHandler, chan types.Order) {
	orderDeliveredPubChan := pubsub.StartPublisher(pubsub.OrderDeliveredDiscoveryPort)
	newOrders := make(chan types.Order)

	oh := OrderHandler{elev: elev}

	var log = logrus.New()

	go func() {
		oh.delayedCounter.Start(counterDelay*time.Millisecond, tickInterval*time.Millisecond)
		defer oh.delayedCounter.Stop()

		for {
			select {
			case order := <-newOrders:
				if len(oh.orders) == 0 {
					oh.delayedCounter.Reset()
				}
				// Set next goal
				oh.orders = append(oh.orders, order)
				nextGoal, err := getNextGoal(oh.orders, oh.elev)
				utils.OkOrPanic(err)
				utils.LogOrder(log, moduleName, "Set next goal", nextGoal)
				currentGoals <- nextGoal
			case arrival := <-arrivals:
				// Delete corresponding order
				for i, v := range oh.orders {
					if utils.OrdersEqual(v, arrival) {
						oh.orders = append(oh.orders[:i], oh.orders[i+1:]...)
						utils.LogOrder(log, moduleName, "Deleted Order", arrival)
						break
					}
				}

				oh.delayedCounter.Reset()

				// Publish order delivered
				arrivalJson, err := json.Marshal(arrival)
				utils.OkOrPanic(err)
				go func() {
					timeout := time.After(1000 * time.Millisecond)
					for {
						select {
						case <-timeout:
							return
						case <-time.After(100 * time.Millisecond):
							orderDeliveredPubChan <- arrivalJson
						}
					}
				}()

				// Set next goal
				if len(oh.orders) > 0 {
					nextGoal, err := getNextGoal(oh.orders, oh.elev)
					utils.LogOrder(log, moduleName, "Set next goal", nextGoal)
					utils.OkOrPanic(err)
					currentGoals <- nextGoal
				}
			}
		}
	}()
	return &oh, newOrders
}

// GetPrice uses calcPriceFromQueue to calculate the price of the call argument, and adds a penalty to order delivery
// delay using the delayed counter.
func (oh *OrderHandler) GetPrice(call types.Call) int {
	price, err := calcPriceFromQueue(types.Order{Call: call}, oh.orders, oh.elev.GetPos(), oh.elev.GetDir())
	utils.OkOrPanic(err)
	if len(oh.orders) > 0 {
		count := <-oh.delayedCounter.Count
		price += int(deliveryDelayWeight * count)
	}
	return price
}

// getNextGoal finds the next goal floor by sorting the order list and picking out the first element.
func getNextGoal(orders []types.Order, elev ElevInterface) (types.Order, error) {
	ordersCopy := make([]types.Order, len(orders))
	copy(ordersCopy, orders)
	sortedOrders, err := sortOrders(ordersCopy, elev.GetPos(), elev.GetDir())
	if err != nil {
		return types.Order{}, err
	}
	return sortedOrders[0], nil
}
