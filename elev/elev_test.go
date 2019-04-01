package elev

import (
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/orders"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	elev := elev{}
	err := elev.Init(elevServerHost, numElevFloors)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStartElevController(t *testing.T) {
	goalArrivals := make(chan types.Order)
	currentGoals := make(chan types.Order)
	floorArrivals := make(chan int)
	go elevio.PollFloorSensor(floorArrivals)

	_ = StartElevController(goalArrivals, currentGoals, floorArrivals)

	firstOrder := types.Order{Call: types.Call{Type: types.Hall, Floor: 3, Dir: types.Down}}
	secondOrder := types.Order{Call: types.Call{Type: types.Cab, Floor: 2, Dir: types.InvalidDir}}
	currentGoals <- firstOrder
	arrived := <-goalArrivals
	if !utils.OrdersEqual(firstOrder, arrived) {
		t.Fatal("Orders not equal")
	}
	currentGoals <- secondOrder
	arrived = <-goalArrivals
	if !utils.OrdersEqual(secondOrder, arrived) {
		t.Fatal("Orders not equal")
	}
}

func TestGoalOverride(t *testing.T) {
	goalArrivals := make(chan types.Order)
	currentGoals := make(chan types.Order)
	floorArrivals := make(chan int)
	go elevio.PollFloorSensor(floorArrivals)

	_ = StartElevController(goalArrivals, currentGoals, floorArrivals)

	firstOrder := types.Order{Call: types.Call{Type: types.Hall, Floor: 3, Dir: types.Down}}
	secondOrder := types.Order{Call: types.Call{Type: types.Cab, Floor: 0, Dir: types.InvalidDir}}

	currentGoals <- firstOrder
	time.Sleep(2000 * time.Millisecond)
	currentGoals <- secondOrder

	arrived := <-goalArrivals
	if !utils.OrdersEqual(secondOrder, arrived) {
		t.Fatal("Orders not equal")
	}
}
