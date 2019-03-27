package utils

import (
	"github.com/sigtot/sanntid/types"
	"github.com/sirupsen/logrus"
)

const logString = "%-15s%s"

// LogBid prints a bid to the terminal
func LogBid(log *logrus.Logger, moduleName string, info string, bid types.Bid) {
	log.WithFields(logrus.Fields{
		"type":  bid.Call.Type,
		"floor": bid.Call.Floor,
		"dir":   bid.Call.Dir,
		"price": bid.Price,
		"id":    bid.ElevatorID,
	}).Infof(logString, moduleName, info)
}

// LogAck prints an ack to the terminal
func LogAck(log *logrus.Logger, moduleName string, info string, ack types.Ack) {
	LogBid(log, moduleName, info, ack.Bid)
}

// LogCall prints a call to the terminal
func LogCall(log *logrus.Logger, moduleName string, info string, call types.Call) {
	log.WithFields(logrus.Fields{
		"type":  call.Type,
		"floor": call.Floor,
		"dir":   call.Dir,
		"id":    call.ElevatorID,
	}).Infof(logString, moduleName, info)
}

// LogOrder prints an order to the terminal
func LogOrder(log *logrus.Logger, moduleName string, info string, order types.Order) {
	LogCall(log, moduleName, info, order.Call)
}

// Log prints a general message to the terminal
func Log(log *logrus.Logger, moduleName string, info string) {
	log.Infof(logString, moduleName, info)
}
