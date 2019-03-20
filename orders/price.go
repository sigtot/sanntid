package orders

import (
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
	"math"
)

const communityWeight = 1
const individualWeight = 0.5
const waitWeight = 3
const travelWeight = 1

func calcPriceFromQueue(newOrder types.Order, orders []types.Order, position float64, dir elevio.MotorDirection) (int, error) {
	sortedOrders, err := sortOrders(orders, position, dir)
	if err != nil {
		return -1, err
	}
	sortedOrders = removeDupesSorted(sortedOrders)

	newSortedOrders, err := sortOrders(append(sortedOrders, newOrder), position, dir)
	if err != nil {
		return -1, err
	}
	newSortedOrders = removeDupesSorted(newSortedOrders)

	newOrderIndex := findOrderIndex(newOrder, newSortedOrders)
	numOrdersAfter := len(newSortedOrders) - (newOrderIndex + 1)
	communityCost := (calcTotalQueueCost(newSortedOrders, position) - calcTotalQueueCost(sortedOrders, position)) * numOrdersAfter
	individualCost := calcTotalQueueCost(newSortedOrders[:newOrderIndex+1], position)
	return int(communityWeight*float64(communityCost) + individualWeight*float64(individualCost) + 0.5), nil
}

// Should take in sorted and unique orders for correct behaviour
func calcTotalQueueCost(orders []types.Order, position float64) int {
	cost := math.Abs(position - float64(orders[0].Floor))
	for i := 1; i < len(orders); i++ {
		cost += math.Abs(float64(orders[i].Floor)-float64(orders[i-1].Floor)) * travelWeight
		if orders[i].Floor != orders[i-1].Floor {
			cost += waitWeight
		}
	}
	return int(cost + 0.5)
}

func removeDupesSorted(orders []types.Order) (uniques []types.Order) {
	uniques = append(uniques, orders[0])
	for i := 1; i < len(orders); i++ {
		if !ordersEqual(orders[i], orders[i-1]) {
			uniques = append(uniques, orders[i])
		}
	}
	return uniques
}

func ordersEqual(order1 types.Order, order2 types.Order) bool {
	return order1.Dir == order2.Dir && order1.Type == order2.Type && order1.Floor == order2.Floor
}

// Returns index of needle in haystack if it exists, otherwise returns -1.
func findOrderIndex(needle types.Order, haystack []types.Order) int {
	for i, v := range haystack {
		if ordersEqual(needle, v) {
			return i
		}
	}
	return -1
}
