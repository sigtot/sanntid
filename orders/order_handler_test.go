package orders

import (
	"encoding/json"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/pubsub"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"log"
	"testing"
	"time"
)

type MockElevatorController struct {
	dir elevio.MotorDirection
	pos float64
}

func (mockElev MockElevatorController) GetDir() elevio.MotorDirection {
	return mockElev.dir
}

func (mockElev MockElevatorController) GetPos() float64 {
	return mockElev.pos
}

func TestOrderHandler(t *testing.T) {
	arrivals := make(chan types.Order)
	currentGoals := make(chan types.Order)

	mockElev := MockElevatorController{dir: elevio.MdUp, pos: 2.0}

	orderDeliveredSubChan, _ := pubsub.StartSubscriber(pubsub.OrderDeliveredDiscoveryPort, "order del")

	_, newOrders := StartOrderHandler(currentGoals, arrivals, mockElev)

	time.Sleep(500 * time.Millisecond)

	newOrder := types.Order{Call: types.Call{Type: types.Hall, Floor: 2, Dir: types.Down}}
	newOrders <- newOrder

	// Simulate elev receiving new goal
	var currentGoal types.Order
	select {
	case currentGoal = <-currentGoals:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for goal")
	}

	newerOrder := types.Order{Call: types.Call{Type: types.Hall, Floor: 3, Dir: types.Down}}
	newOrders <- newerOrder
	select {
	case currentGoal = <-currentGoals:
		if !utils.OrdersEqual(currentGoal, newerOrder) {
			t.Fatal("Order and current goal does not match")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for goal")
	}

	// Simulate elev arriving at goal and notifying order handler
	arrivals <- currentGoal

	// Check that we received delivered order thing
	js := <-orderDeliveredSubChan
	var orderDelivered types.Order
	err := json.Unmarshal(js, &orderDelivered)
	utils.OkOrPanic(err)
	if !utils.OrdersEqual(orderDelivered, newerOrder) {
		log.Fatal("Delivered order not equal to newer order")
	}

	select {
	case currentGoal = <-currentGoals:
		if !utils.OrdersEqual(currentGoal, newOrder) {
			t.Fatal("Order and current goal does not match")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for goal")
	}
}
