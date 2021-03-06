/*
Package elev implements a simple elevator controller that directs the elevator
to a goal floor and handles arrival at goal floor.
*/
package elev

import (
	"errors"
	"fmt"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Door open wait time in milliseconds
const doorOpenWaitTime = 3000

// Init timeout time in milliseconds
const initTimeout = 3000

const elevServerHost = "localhost"

const numElevFloors = 4

const moduleName = "ELEV"
const logString = "%-15s%s"

type elev struct {
	dir      elevio.MotorDirection
	pos      float64
	goal     types.Order
	doorOpen bool
}

// StartElevController initializes the elevator controller and starts a go-routine that
// responds to new goals on currentGoals and announces goal arrival at goalArrival.
func StartElevController(
	goalArrivals chan<- types.Order,
	currentGoals <-chan types.Order,
	floorArrivals <-chan int,
	elevPort int,
	quit <-chan int,
	wg *sync.WaitGroup) *elev {
	var log = logrus.New()
	atGoal := make(chan int, 1024)

	elev := elev{}
	elevServerAddr := fmt.Sprintf("%s:%d", elevServerHost, elevPort)
	err := elev.Init(elevServerAddr, numElevFloors, floorArrivals)
	utils.OkOrPanic(err)

	log.WithFields(logrus.Fields{
		"addr": elevServerAddr,
	}).Infof(logString, moduleName, "Successfully initiated elev server")

	var startAgain <-chan time.Time

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case elev.goal = <-currentGoals:
				// Set direction to deliver new goal order
				newGoalDir, updateDir, err := goalDir(elev.goal, elev.pos)
				utils.OkOrPanic(err)
				if updateDir == true {
					elev.dir = newGoalDir
				}

				if elev.atGoal() {
					atGoal <- 1
				} else if !elev.doorOpen {
					elev.start()
				}
			case <-atGoal:
				// Stop elevator, open doors and announce arrival
				elev.stop()
				elev.doorOpen = true
				elevio.SetDoorOpenLamp(true)
				startAgain = time.After(doorOpenWaitTime * time.Millisecond)
				goalArrivals <- elev.goal
				utils.Log(log, moduleName, "Opened doors")
			case floorArrival := <-floorArrivals:
				if floorArrival < 0 {
					// Between floors
					isWholeNumber := float64(int(elev.pos)) == elev.pos
					if isWholeNumber {
						elev.setPos(elev.pos + 0.5*float64(elev.dir))
					}
				} else {
					// At floor
					elev.setPos(float64(floorArrival))
					if int(elev.pos) == elev.goal.Floor {
						atGoal <- 1
					}
				}
			case <-startAgain:
				// Close doors and start elevator again
				elevio.SetDoorOpenLamp(false)
				elev.doorOpen = false
				if !elev.atGoal() {
					elev.start()
				}
				utils.Log(log, moduleName, "Closed doors")
			case <-quit:
				elevio.SetMotorDirection(elevio.MdStop)
				elevio.SetDoorOpenLamp(false)
				utils.Log(log, moduleName, "Turned off motor and closed door")
				return
			}
		}
	}()
	return &elev
}

// Init initializes elevio and moves the elevator down to a floor in order to determine the position
func (elev *elev) Init(addr string, numFloors int, floorArrivals <-chan int) error {
	elevio.Init(addr, numFloors)

	elevio.SetMotorDirection(elevio.MdDown)
	elev.dir = elevio.MdDown

	defer elev.stop()
	timeout := time.After(initTimeout * time.Millisecond)
L:
	for {
		select {
		case floor := <-floorArrivals:
			elev.setPos(float64(floor))
			break L
		case <-timeout:
			return errors.New("failed to reach floor within timeout")
		}
	}

	return nil
}

// goalDir calculates the new goal direction from the goal order argument and the current position
func goalDir(goal types.Order, pos float64) (dir elevio.MotorDirection, updateDir bool, err error) {
	if float64(goal.Floor) > pos {
		return elevio.MdUp, true, nil
	} else if float64(goal.Floor) < pos {
		return elevio.MdDown, true, nil
	} else if goal.Type == types.Hall {
		if dir, err := utils.OrderDir2MDDir(goal.Dir); err == nil {
			return dir, true, nil
		}
		return dir, false, err
	}
	return elevio.MdDown, false, nil
}

func (elev *elev) stop() {
	elevio.SetMotorDirection(elevio.MdStop)
}

func (elev *elev) start() {
	elevio.SetMotorDirection(elev.dir)
}

func (elev *elev) setPos(pos float64) {
	elev.pos = pos
	isWholeNumber := float64(int(elev.pos)) == elev.pos
	if isWholeNumber {
		elevio.SetFloorIndicator(int(pos))
	}
}

func (elev *elev) atGoal() bool {
	return int(2*elev.pos) == 2*elev.goal.Floor
}

func (elev *elev) GetDir() elevio.MotorDirection {
	return elev.dir
}

func (elev *elev) GetPos() float64 {
	return elev.pos
}
