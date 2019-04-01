package utils

import (
	"errors"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
)

// OrderDir2MDDir converts an order direction to a motor direction
func OrderDir2MDDir(orderDir types.Direction) (mdDir elevio.MotorDirection, err error) {
	if orderDir == types.Up {
		return elevio.MdUp, nil
	}
	if orderDir == types.Down {
		return elevio.MdDown, nil
	}
	return elevio.MdStop, errors.New("conversion from invalid order direction to motor direction")
}

func OkOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
