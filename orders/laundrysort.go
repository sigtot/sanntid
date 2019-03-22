package orders

import (
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
)

func sortOrders(orders []types.Order, position float64, dir elevio.MotorDirection) (sorted []types.Order, err error) {
	// Choose a starting direction if elevator standing still
	if dir == elevio.MdStop {
		dir = findStartDir(orders, int(position))
	}

	// Iterate over elevator cycle
	floor := roundPositionInDirection(position, dir)
	startFloor := floor
	startDir := dir
	homeHitCount := 0
	for {
		if floor == startFloor && dir == startDir {
			homeHitCount++
		}
		if homeHitCount >= 2 {
			break
		}

		for i := 0; i < len(orders); i++ {
			order := orders[i]
			if order.Floor == floor {
				if order.Type == types.Cab {
					sorted = append(sorted, order)
					orders = append(orders[:i], orders[i+1:]...)
					i--
				} else if mdDir, e := utils.OrderDir2MDDir(order.Dir); e == nil {
					if mdDir == dir {
						sorted = append(sorted, order)
						orders = append(orders[:i], orders[i+1:]...)
						i--
					}
				} else {
					err = e
				}
			}
		}

		if floor >= topFloor && dir == elevio.MdUp {
			dir = elevio.MdDown
		} else if floor <= bottomFloor && dir == elevio.MdDown {
			dir = elevio.MdUp
		} else {
			floor += int(dir)
		}
	}

	i := 0
	for j := i + 1; j <= len(sorted); j++ {
		if j == len(sorted) || sorted[j].Floor != sorted[j-1].Floor {
			for k := i; k < j; k++ {
				if sorted[k].Type == types.Cab {
					temp := sorted[i]
					sorted[i] = sorted[k]
					sorted[k] = temp
					i++
				}
			}
			i = j
		}
	}
	return sorted, err
}

func roundPositionInDirection(position float64, dir elevio.MotorDirection) (floor int) {
	if dir == elevio.MdDown {
		return int(position)
	}
	return int(position + 0.5)
}

func findStartDir(orders []types.Order, floor int) elevio.MotorDirection {
	for d := 0; d < numFloors; d += 1 {
		for _, order := range orders {
			if order.Floor == floor+d {
				return elevio.MdUp
			}
			if order.Floor == floor-d {
				return elevio.MdDown
			}
		}
	}
	return elevio.MdUp
}