package orders

import (
	"errors"
	"fmt"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
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
	for {
		for i := 0; i < len(orders); i++ {
			order := orders[i]
			if order.Floor == floor {
				if order.Type == types.Cab {
					sorted = append(sorted, order)
					orders = append(orders[:i], orders[i+1:]...)
					i--
				} else if mdDir, e := orderDir2MDDir(order.Dir); e == nil {
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
		fmt.Println()
		floor += int(dir)
		if floor == topFloor || floor == bottomFloor {
			dir = revDir(dir)
		}
		if floor == startFloor && dir == startDir {
			break
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

func orderDir2MDDir(orderDir types.Direction) (mdDir elevio.MotorDirection, err error) {
	if orderDir == types.Up {
		return elevio.MdUp, nil
	}
	if orderDir == types.Down {
		return elevio.MdDown, nil
	}
	return elevio.MdStop, errors.New("conversion from invalid order direction to motor direction")
}

func revDir(dir elevio.MotorDirection) elevio.MotorDirection {
	return elevio.MotorDirection(-1 * int(dir))
}
