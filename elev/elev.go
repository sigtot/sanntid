package elev

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"sync"
	"time"
)

// Door open wait time in milliseconds
const doorOpenWaitTime = 3000

// It's timeout time
const initTimeoutTime = 3000

const elevServerAddr = "localhost:15657"

const numElevFloors = 4

type Elev struct {
	dir      elevio.MotorDirection
	pos      float64
	goal     types.Order
	doorOpen bool
	doorMu   sync.Mutex
}

func StartElevController(goalArrivals chan<- types.Order, currentGoals <-chan types.Order, floorArrivals <-chan int) *Elev {
	elev := Elev{}
	err := elev.Init(elevServerAddr, numElevFloors)
	if err != nil {
		panic(err)
	}

	var log = logrus.New()

	atGoal := make(chan int, 1024)

	var startAgain <-chan time.Time
	go func() {
		for {
			select {
			case elev.goal = <-currentGoals:
				// TODO: Think hard about this
				elev.doorMu.Lock()
				if elev.atGoal() {
					atGoal <- 1
				} else if !elev.doorOpen {
					elev.setDir(elev.getGoalDir())
				}
				elev.doorMu.Unlock()
			case <-atGoal:
				elev.doorMu.Lock()
				elev.setDir(elevio.MdStop)
				elev.doorOpen = true
				elev.doorMu.Unlock()
				startAgain = time.After(doorOpenWaitTime * time.Millisecond)
				elevio.SetDoorOpenLamp(true)
				goalArrivals <- elev.goal
				utils.Log(log, "ELEV", "Opened doors")
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
				elevio.SetDoorOpenLamp(false)
				elev.doorMu.Lock()
				elev.doorOpen = false
				elev.doorMu.Unlock()
				elev.setDir(elev.getGoalDir())
				utils.Log(log, "ELEV", "Closed doors")
			}
		}
	}()
	return &elev
}

func (elev *Elev) getGoalDir() elevio.MotorDirection {
	if float64(elev.goal.Floor) > elev.pos {
		return elevio.MdUp
	} else if float64(elev.goal.Floor) < elev.pos {
		return elevio.MdDown
	}
	return elevio.MdStop
}

func (elev *Elev) Init(addr string, numFloors int) error {
	var log = logrus.New()
	elevio.Init(addr, numFloors)
	floorArrivals := make(chan int)
	go elevio.PollFloorSensor(floorArrivals)

	elev.setDir(elevio.MdDown)
	defer elev.setDir(elevio.MdStop)
	timeout := time.After(initTimeoutTime * time.Millisecond)
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

	utils.Log(log, "ELEV", "Elev init successful")

	return nil
}

func (elev *Elev) setDir(direction elevio.MotorDirection) {
	elev.dir = direction
	elevio.SetMotorDirection(direction)
}

func (elev *Elev) setPos(pos float64) {
	elev.pos = pos
	isWholeNumber := float64(int(elev.pos)) == elev.pos
	if isWholeNumber {
		elevio.SetFloorIndicator(int(pos))
	}
}

func (elev *Elev) atGoal() bool {
	return int(2*elev.pos) == 2*elev.goal.Floor
}

func (elev *Elev) GetDir() elevio.MotorDirection {
	return elev.dir
}

func (elev *Elev) GetPos() float64 {
	return elev.pos
}
