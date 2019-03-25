package orders

import (
	"encoding/json"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	"time"
)

const numFloors = 4
const topFloor = numFloors - 1
const bottomFloor = 0
const moduleName = "ORDER HANDLER"

const counterDelay = 12000
const tickInterval = 1000
const deliveryDelayWeight = 1

type OrderHandler struct {
	orders         []types.Order
	delayedCounter DelayedCounter
	elev           ElevInterface
}

type ElevInterface interface {
	GetDir() elevio.MotorDirection
	GetPos() float64
}

func StartOrderHandler(
	currentGoals chan types.Order,
	arrivals chan types.Order,
	elev ElevInterface) (*OrderHandler, chan types.Order) {
	orderDeliveredPubChan := publish.StartPublisher(pubsub.OrderDeliveredDiscoveryPort)
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
				okOrPanic(err)
				utils.LogOrder(log, moduleName, "Set next goal", nextGoal)
				currentGoals <- nextGoal
			case arrival := <-arrivals:
				// Delete corresponding order
				for i, v := range oh.orders {
					if OrdersEqual(v, arrival) {
						oh.orders = append(oh.orders[:i], oh.orders[i+1:]...)
						utils.LogOrder(log, moduleName, "Deleted Order", arrival)
						break
					}
				}

				oh.delayedCounter.Reset()

				// Publish order delivered
				arrivalJson, err := json.Marshal(arrival)
				okOrPanic(err)
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
					okOrPanic(err)
					currentGoals <- nextGoal
				}
			}
		}
	}()
	return &oh, newOrders
}

func (oh *OrderHandler) GetPrice(order types.Call) int {
	price, err := calcPriceFromQueue(types.Order{Call: order}, oh.orders, oh.elev.GetPos(), oh.elev.GetDir())
	okOrPanic(err)
	if len(oh.orders) > 0 {
		count := <-oh.delayedCounter.Count
		price += int(deliveryDelayWeight * count)
	}
	return price
}

func getNextGoal(orders []types.Order, elev ElevInterface) (types.Order, error) {
	ordersCopy := make([]types.Order, len(orders))
	copy(ordersCopy, orders)
	sortedOrders, err := sortOrders(ordersCopy, elev.GetPos(), elev.GetDir())
	if err != nil {
		return types.Order{}, err
	}
	return sortedOrders[0], nil
}

func okOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
