/*
Package buttons contains logic for listening to new calls, translating them and sending them to a seller.
*/
package buttons

import (
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/mac"
	"github.com/sigtot/sanntid/types"
)

// StartButtonHandler starts a go-routine that listens for button events on the buttonEvents channel,
// and translates the received event to a call type and sends it on the callsForSale channel, which is then received
// by a seller.
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
