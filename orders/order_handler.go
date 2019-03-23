package orders

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/pubsub/publish"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
)

const numFloors = 4 // TODO: Move this maybe
const topFloor = numFloors - 1
const bottomFloor = 0

type OrderHandler struct {
	orders []types.Order
	elev   ElevInterface
}

type ElevInterface interface {
	GetDir() elevio.MotorDirection
	GetPos() float64
}

var log = logrus.New()

func StartOrderHandler(
	currentGoals chan types.Order,
	arrivals chan types.Order,
	elev ElevInterface) (*OrderHandler, chan types.Order) {
	orderDeliveredPubChan := publish.StartPublisher(pubsub.OrderDeliveredDiscoveryPort)
	oh := OrderHandler{elev: elev}
	newOrders := make(chan types.Order)

	var log = logrus.New()

	go func() {
		for {
			select {
			case order := <-newOrders:
				// Set next goal
				oh.orders = append(oh.orders, order)
				nextGoal, err := getNextGoal(oh.orders, oh.elev)
				okOrPanic(err)
				utils.LogOrder(log, "ORDER HANDLER", "Next Goal Order", nextGoal)
				currentGoals <- nextGoal
			case arrival := <-arrivals:
				// Delete corresponding order
				for i, v := range oh.orders {
					if OrdersEqual(v, arrival) {
						oh.orders = append(oh.orders[:i], oh.orders[i+1:]...)
						utils.LogOrder(log, "ORDER HANDLER", "Deleting Order", arrival)
						break
					}
				}
				// Publish order delivered
				arrivalJson, err := json.Marshal(arrival)
				okOrPanic(err)
				orderDeliveredPubChan <- arrivalJson

				// Set next goal
				if len(oh.orders) > 0 {
					nextGoal, err := getNextGoal(oh.orders, oh.elev)
					utils.LogOrder(log, "ORDER HANDLER", "Next Goal Order", nextGoal)
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
	return price
	// TODO: Implement base price
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
