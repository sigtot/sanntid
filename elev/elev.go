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

// It's timeout time
const initTimeoutTime = 3000

const elevServerHost = "localhost"

const numElevFloors = 4

const moduleName = "ELEV"

type Elev struct {
	dir      elevio.MotorDirection
	moving   bool
	pos      float64
	goal     types.Order
	doorOpen bool
}

func StartElevController(goalArrivals chan<- types.Order, currentGoals <-chan types.Order, floorArrivals <-chan int, elevPort int, quit <-chan int, wg *sync.WaitGroup) *Elev {
	var log = logrus.New()

	elev := Elev{}
	elevServerAddr := fmt.Sprintf("%s:%d", elevServerHost, elevPort)
	err := elev.Init(elevServerAddr, numElevFloors, floorArrivals)
	if err != nil {
		panic(err)
	}
	log.WithFields(logrus.Fields{
		"addr": elevServerAddr,
	}).Infof("%-15s %s", moduleName, "Successfully initiated elev server")

	atGoal := make(chan int, 1024)

	wg.Add(1)

	var startAgain <-chan time.Time
	go func() {
		defer wg.Done()

		for {
			select {
			case elev.goal = <-currentGoals:
				if newGoalDir, updateDir, err := goalDir(elev.goal, elev.pos); updateDir == true && err == nil {
					elev.dir = newGoalDir
				} else if err != nil {
					panic(err)
				}
				// TODO: Think hard about this
				if elev.atGoal() {
					atGoal <- 1
				} else if !elev.doorOpen {
					elev.start()
				}
			case <-atGoal:
				elev.stop()
				elev.doorOpen = true
				startAgain = time.After(doorOpenWaitTime * time.Millisecond)
				elevio.SetDoorOpenLamp(true)
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

func (elev *Elev) Init(addr string, numFloors int, floorArrivals <-chan int) error {
	elevio.Init(addr, numFloors)

	elev.moving = true
	elevio.SetMotorDirection(elevio.MdDown)
	elev.dir = elevio.MdDown

	defer elev.stop()
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

	return nil
}

func goalDir(goal types.Order, pos float64) (dir elevio.MotorDirection, updateDir bool, err error) {
	if float64(goal.Floor) > pos {
		return elevio.MdUp, true, nil
	} else if float64(goal.Floor) < pos {
		return elevio.MdDown, true, nil
	} else if goal.Type == types.Hall {
		if dir, err := utils.OrderDir2MDDir(goal.Dir); err == nil {
			return dir, true, nil
		} else {
			return dir, false, err
		}
	} else {
		return elevio.MdDown, false, nil
	}
}

func (elev *Elev) stop() {
	elev.moving = false
	elevio.SetMotorDirection(elevio.MdStop)
}

func (elev *Elev) start() {
	elev.moving = true
	elevio.SetMotorDirection(elev.dir)
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
