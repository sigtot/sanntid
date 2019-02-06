package main

import (
	"github.com/sigtot/sanntid/broker"
	"github.com/sigtot/sanntid/order"
)

func main() {
	// Start selling
	ordChan := make(chan order.Order)
	broker.StartSelling(ordChan)

	// Start buyer

	// Receive order from panel
	ord := order.Order{Floor: 3, Dir: order.Down}

	ordChan <- ord
}
