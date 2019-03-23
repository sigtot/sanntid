package orders

import (
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
	"testing"
)

func TestCalcPriceFromQueue(t *testing.T) {
	orders := []types.Order{
		{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
		{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
		{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
	}

	newOrder := types.Order{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}}
	price, err := calcPriceFromQueue(newOrder, orders, 3.0, elevio.MdDown)
	expectedCost := 20
	if err != nil {
		log.Fatal(err.Error())
	}
	if price != expectedCost {
		log.Fatalf("Fail: Got price %d but expected %d\n", price, expectedCost)
	}

	// This order already exists, and so should only give individual price
	oldOrder := types.Order{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}}
	price, err = calcPriceFromQueue(oldOrder, orders, 3.0, elevio.MdDown)
	expectedCost = 4 // (1 travel cost + 3 wait cost) * 1 individualWeight = 4
	if err != nil {
		log.Fatal(err.Error())
	}
	if price != expectedCost {
		log.Fatalf("Fail: Got price %d but expected %d\n", price, expectedCost)
	}

	// Order where we are
	sameFloorOrder := types.Order{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}}
	price, err = calcPriceFromQueue(sameFloorOrder, orders, 2.0, elevio.MdDown)
	expectedCost = 0
	if err != nil {
		log.Fatal(err.Error())
	}
	if price != expectedCost {
		log.Fatalf("Fail: Got price %d but expected %d\n", price, expectedCost)
	}
}

func TestCalcTotalQueueCost(t *testing.T) {
	orders := []types.Order{
		{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
		{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
		{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
	}

	dir := elevio.MdUp
	pos := 3.0
	orders, err := sortOrders(orders, pos, dir)
	if err != nil {
		log.Fatal(err.Error())
	}
	orders = removeDupesSorted(orders)
	cost := calcTotalQueueCost(orders, pos)
	expectedCost := 14
	if cost != expectedCost {
		log.Fatalf("Got cost %d but expected %d\n", cost, expectedCost)
	}
}

func TestCalcPriceFromEmptyQueue(t *testing.T) {
	var orders []types.Order

	newOrder := types.Order{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}}
	price, err := calcPriceFromQueue(newOrder, orders, 3.0, elevio.MdDown)
	expectedCost := 2
	if err != nil {
		log.Fatal(err.Error())
	}
	if price != expectedCost {
		log.Fatalf("Fail: Got price %d but expected %d\n", price, expectedCost)
	}
}

func TestRemoveDupesFromEmptyQueue(t *testing.T) {
	var orders []types.Order
	_ = removeDupesSorted(orders)
}
