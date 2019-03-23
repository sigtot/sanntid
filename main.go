package main

import (
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/buttons"
	"github.com/sigtot/sanntid/buyer"
	"github.com/sigtot/sanntid/elev"
	"github.com/sigtot/sanntid/indicators"
	"github.com/sigtot/sanntid/orders"
	"github.com/sigtot/sanntid/seller"
	"github.com/sigtot/sanntid/types"
)

func main() {
	goalArrivals := make(chan types.Order)
	currentGoals := make(chan types.Order)
	floorArrivals := make(chan int)
	go elevio.PollFloorSensor(floorArrivals)
	elevator := elev.StartElevController(goalArrivals, currentGoals, floorArrivals)

	callsForSale := make(chan types.Call)
	buttonEvents := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(buttonEvents)
	buttons.StartButtonHandler(buttonEvents, callsForSale)

	indicators.StartIndicatorHandler()

	oh, newOrders := orders.StartOrderHandler(currentGoals, goalArrivals, elevator)

	buyer.StartBuying(oh, newOrders)

	seller.StartSelling(callsForSale) //Go this
}
