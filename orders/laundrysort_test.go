package orders

import (
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

		sortedOrders, err := sortOrders(orders, 1, elevio.MdUp)
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

		sortedOrders, err := sortOrders(orders, 1, elevio.MdUp)
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

		sortedOrders, err := sortOrders(orders, 3, elevio.MdDown)
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
	_, err := sortOrders(orders, 2.0, elevio.MdDown)
	if err != nil {
		log.Fatal(err.Error())
	}
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
