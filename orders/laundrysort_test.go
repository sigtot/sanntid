package orders

import (
	"fmt"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
	"log"
	"math/rand"
	"reflect"
	"testing"
)

func TestSortOrders(t *testing.T) {
	for k := 0; k < 100; k++ {
		orders := []types.Order{
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 1}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 1}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 3}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 3}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 0}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 0}},
		}

		orderedOrders := []types.Order{
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 1}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 1}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 3}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 3}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 1}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 0}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 0}},
		}

		orders = scramble(orders)

		sortedOrders, err := SortOrders(orders, 1, elevio.MdUp)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !reflect.DeepEqual(sortedOrders, orderedOrders) {
			t.Logf("Sorted wrong in iteration %d: Expected:\n", k)
			t.Logf("%+v\n", orderedOrders)
			t.Logf("but got\n")
			t.Logf("%+v\n", sortedOrders)
			t.FailNow()
		}
	}
}

func TestSortOrdersNice(t *testing.T) {
	for k := 0; k < 100; k++ {
		orders := []types.Order{
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
		}

		orderedOrders := []types.Order{
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
		}

		orders = scramble(orders)

		sortedOrders, err := SortOrders(orders, 1.5, elevio.MdUp)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !reflect.DeepEqual(sortedOrders, orderedOrders) {
			t.Logf("Sorted wrong in iteration %d: Expected:\n", k)
			t.Logf("%+v\n", orderedOrders)
			t.Logf("but got\n")
			t.Logf("%+v\n", sortedOrders)
			t.FailNow()
		}
	}
}

func TestSortDownwards(t *testing.T) {
	for k := 0; k < 100; k++ {
		orders := []types.Order{
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
		}

		orderedOrders := []types.Order{
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 3}},
			{Call: types.Call{Type: types.Hall, Dir: types.Down, Floor: 2}},
			{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 0}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
			{Call: types.Call{Type: types.Hall, Dir: types.Up, Floor: 2}},
		}

		orders = scramble(orders)

		sortedOrders, err := SortOrders(orders, 3, elevio.MdDown)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !reflect.DeepEqual(sortedOrders, orderedOrders) {
			t.Logf("Sorted wrong in iteration %d: Expected:\n", k)
			t.Logf("%+v\n", orderedOrders)
			t.Logf("but got\n")
			t.Logf("%+v\n", sortedOrders)
			t.FailNow()
		}
	}
}

func TestSortEmptyQueue(t *testing.T) {
	var orders []types.Order
	_, err := SortOrders(orders, 2.0, elevio.MdDown)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// SortOrders will sort orders that are closest to the current position first.
func ExampleSortOrders() {
	orders := []types.Order{
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 1}},
		{Call: types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: 2}},
	}

	pos := 3.0
	dir := elevio.MdDown

	fmt.Printf("Before: %+v\n", orders)
	if sortedOrders, err := SortOrders(orders, pos, dir); err == nil {
		fmt.Printf("After: %+v\n", sortedOrders)
	}
	// Output:
	// Before: [{Call:{Type:0 Floor:1 Dir:-1 ElevatorID:}} {Call:{Type:0 Floor:2 Dir:-1 ElevatorID:}}]
	// After: [{Call:{Type:0 Floor:2 Dir:-1 ElevatorID:}} {Call:{Type:0 Floor:1 Dir:-1 ElevatorID:}}]
}

func scramble(orders []types.Order) []types.Order {
	for i := len(orders) - 1; i > 0; i-- {
		j := rand.Intn(i)
		temp := orders[i]
		orders[i] = orders[j]
		orders[j] = temp
	}
	return orders
}
