package elev

import (
	"fmt"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/orders"
	"github.com/sigtot/sanntid/types"
	"testing"
)

func TestInit(t *testing.T) {
	elev := Elev{}
	err := elev.Init(elevServerAddr, numElevFloors)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPollFloor(t *testing.T) {
	elevio.Init(elevServerAddr, numElevFloors)
	floorChan := make(chan int)
	go elevio.PollFloorSensor(floorChan)
	for {
		floor := <-floorChan
		fmt.Printf("Floor: %d\n", floor)
	}
}

func TestStartElevController(t *testing.T) {
	goalArrivals := make(chan types.Order)
	currentGoals := make(chan types.Order)

	_ = StartElevController(goalArrivals, currentGoals)

	newOrder := types.Order{Call: types.Call{Type: types.Hall, Floor: 2, Dir: types.Up}}
	currentGoals <- newOrder
	arrived := <-goalArrivals
	if !orders.OrdersEqual(newOrder, arrived) {
		t.Fatal("Orders not equal")
	}
}
