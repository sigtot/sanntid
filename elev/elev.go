package elev

import (
	"github.com/sigtot/elevio"
	"time"
)

var direction elevio.MotorDirection
var position float64

type State struct {
	Dir elevio.MotorDirection
	Pos float64
}

func StartElev(goal <-chan int, arrived chan<- int) {
	floorSensorChan := make(chan int)
	elevio.PollFloorSensor(floorSensorChan)
	var goalFloor int

	for {
		select {
		case goalFloor = <-goal:
			direction = calcDirection(position, goalFloor)
			elevio.SetMotorDirection(direction)
		case position = <-floorSensorChan:
			elevio.SetFloorIndicator(int(position))
			if int(position) == goalFloor {
				func() {
					direction = elevio.MD_Stop
					elevio.SetMotorDirection(direction)
					elevio.SetDoorOpenLamp(true)
					defer elevio.SetDoorOpenLamp(false)
					arrived <- goalFloor
					time.Sleep(3 * time.Second)
				}()
			} else if direction == elevio.MD_Up {
				position += 0.5
			} else if direction == elevio.MD_Down {
				position -= 0.5
			}
		}
	}
}

func calcDirection(pos float64, goalFloor int) elevio.MotorDirection {
	return elevio.MD_Stop
}

func GetDirection() elevio.MotorDirection {
	return direction
}

func GetPosition() float64 {
	return position
}
