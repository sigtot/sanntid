package main

import (
	"fmt"
	"github.com/TTK4145/driver-go/elevio"
	"reflect"
	"time"
)

const numFloors = 4

type Order struct {
	Floor      int
	ButtonType elevio.ButtonType
}

func main() {
	var orders []Order

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
			newOrder := Order{
				Floor:      a.Floor,
				ButtonType: a.Button,
			}
			if !OrderExists(orders, newOrder) {
				orders = append(orders, newOrder)
			}
			fmt.Printf("Orders: %x\n", orders)
		case a := <-drvFloors:
			fmt.Printf("%+v\n", a)
			if OrderExists(orders, Order{Floor: a, ButtonType: elevio.BT_Cab}) || OrderExists(orders, Order{Floor: a, ButtonType: MDToBT(d)}) {
				ServeOrder()
				orders = DeleteOrder(orders, a, MDToBT(d))
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

func OrderExists(orders []Order, order Order) bool {
	for i := 0; i < len(orders); i++ {
		if reflect.DeepEqual(orders[i], order) {
			return true
		}
	}
	return false
}

func ServeOrder() {
	elevio.SetMotorDirection(elevio.MD_Stop)
	elevio.SetDoorOpenLamp(true)
	time.Sleep(time.Second * 3)
	elevio.SetDoorOpenLamp(false)
}

func DeleteOrder(orders []Order, floor int, buttonType elevio.ButtonType) []Order {
	for i := 0; i < len(orders); i++ {
		if orders[i].Floor == floor && (orders[i].ButtonType == buttonType || orders[i].ButtonType == elevio.BT_Cab) {
			orders = append(orders[:i], orders[i+1:]...)
			elevio.SetButtonLamp(buttonType, floor, false)
			elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		}
	}
	return orders
}

func MDToBT(motorDir elevio.MotorDirection) elevio.ButtonType {
	return elevio.ButtonType(int(motorDir))
}
