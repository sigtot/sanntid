package elev

import (
	"github.com/TTK4145/driver-go/elevio"
	"time"
)

type State struct {
	Dir elevio.MotorDirection
	Pos float64
}

func StartElev(goal <-chan int, arrived chan<- int, state chan<- State) {
	floorSensorChan := make(chan int)
	stateChange := make(chan int)
	elevio.PollFloorSensor(floorSensorChan)
	var goalFloor int
	var dir elevio.MotorDirection
	var pos float64

	for {
		select {
		case goalFloor = <-goal:
			dir = getDirection(pos, goalFloor)
			elevio.SetMotorDirection(dir)
			stateChange <- 1
		case pos = <-floorSensorChan:
			elevio.SetFloorIndicator(int(pos))
			stateChange <- 1
			if int(pos) == goalFloor {
				func() {
					dir = elevio.MD_Stop
					elevio.SetMotorDirection(dir)
					elevio.SetDoorOpenLamp(true)
					defer elevio.SetDoorOpenLamp(false)
					arrived <- goalFloor
					time.Sleep(3 * time.Second)
				}()
			} else if dir == elevio.MD_Up {
				pos += 0.5
			} else if dir == elevio.MD_Down {
				pos -= 0.5
			}
		case <-stateChange:
			// TODO: Don't do this. Do https://stackoverflow.com/questions/27236827/idiomatic-way-to-make-a-request-response-communication-using-channels
			state <- State{Dir: dir, Pos: pos}
		}
	}
}

func getDirection(pos float64, goalFloor int) elevio.MotorDirection {
	return elevio.MD_Stop
}
