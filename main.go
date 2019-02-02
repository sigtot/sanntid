package main

import (
	"fmt"
	"github.com/TTK4145/driver-go/elevio"
	"reflect"
	"time"
)

const numFloors = 4

type order struct {
	Floor      int
	ButtonType elevio.ButtonType
}

func main() {
	var orders []order

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)

	fmt.Println("Starting elevator")
	elevio.Init("127.0.0.1:15657", numFloors)
	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)

	d := elevio.MD_Up
	elevio.SetMotorDirection(d)

	for {
		select {
		case a := <-drvButtons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			newOrder := order{
				Floor:      a.Floor,
				ButtonType: a.Button,
			}
			if !orderExists(orders, newOrder) {
				orders = append(orders, newOrder)
			}
			fmt.Printf("Orders: %x\n", orders)
		case a := <-drvFloors:
			fmt.Printf("%+v\n", a)
			if orderExists(orders, order{Floor: a, ButtonType: elevio.BT_Cab}) || orderExists(orders, order{Floor: a, ButtonType: mdToBT(d)}) {
				serveOrder()
				orders = deleteOrder(orders, a, mdToBT(d))
				fmt.Printf("Orders: %x\n", orders)
			}
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}

			elevio.SetMotorDirection(d)
		}
	}
}

func orderExists(orders []order, order order) bool {
	for i := 0; i < len(orders); i++ {
		if reflect.DeepEqual(orders[i], order) {
			return true
		}
	}
	return false
}

func serveOrder() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	time.Sleep(time.Second * 3)
	elevio.SetDoorOpenLamp(false)
}

func deleteOrder(orders []order, floor int, buttonType elevio.ButtonType) []order {
	for i := 0; i < len(orders); i++ {
		if orders[i].Floor == floor && (orders[i].ButtonType == buttonType || orders[i].ButtonType == elevio.BT_Cab) {
			orders = append(orders[:i], orders[i+1:]...)
			elevio.SetButtonLamp(buttonType, floor, false)
			elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		}
	}
	return orders
}

func mdToBT(motorDir elevio.MotorDirection) elevio.ButtonType {
	return elevio.ButtonType(int(motorDir))
}
