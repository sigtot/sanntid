package buttons

import (
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/types"
)

func StartButtonHandler(buttonEvents chan elevio.ButtonEvent, callsForSale chan types.Call) {
	ID, err := mac.GetMacAddr()
	if err != nil {
		panic("Could not get mac address")
	}
	go func() {
		for {
			buttonEvent := <-buttonEvents
			var call types.Call
			if buttonEvent.Button == elevio.BtnHallUp {
				call = types.Call{Type: types.Hall, Dir: types.Up, Floor: buttonEvent.Floor}
			} else if buttonEvent.Button == elevio.BtnHallDown {
				call = types.Call{Type: types.Hall, Dir: types.Down, Floor: buttonEvent.Floor}
			} else {
				call = types.Call{Type: types.Cab, Dir: types.InvalidDir, Floor: buttonEvent.Floor, ElevatorID: ID}
			}
			callsForSale <- call
		}
	}()
}
