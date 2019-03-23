package utils

import (
	"github.com/Sirupsen/logrus"
	"github.com/sigtot/sanntid/types"
)

func LogBid(log *logrus.Logger, moduleName string, info string, bid types.Bid) {
	log.WithFields(logrus.Fields{
		"type":  bid.Call.Type,
		"floor": bid.Call.Floor,
		"dir":   bid.Call.Dir,
		"price": bid.Price,
		"id":    bid.ElevatorID,
	}).Infof("%-15s %s", moduleName, info)
}

func LogAck(log *logrus.Logger, moduleName string, info string, ack types.Ack) {
	LogBid(log, moduleName, info, ack.Bid)
}

func LogCall(log *logrus.Logger, moduleName string, info string, call types.Call) {
	log.WithFields(logrus.Fields{
		"type":  call.Type,
		"floor": call.Floor,
		"dir":   call.Dir,
		"id":    call.ElevatorID,
	}).Infof("%-15s %s", moduleName, info)
}

func LogOrder(log *logrus.Logger, moduleName string, info string, order types.Order) {
	LogCall(log, moduleName, info, order.Call)
}

func LogNewSub(log *logrus.Logger, moduleName string, info string, subIP string) {
	log.WithFields(logrus.Fields{
		"IP": subIP,
	}).Infof("%-15s %s", moduleName, info)
}

func Log(log *logrus.Logger, moduleName string, info string) {
	log.Infof("%-15s %s", moduleName, info)
}
