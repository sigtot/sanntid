package buttons

import (
	"fmt"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/types"
	"testing"
)

const elevServerAddr = "localhost:15657"

const numElevFloors = 4

func TestStartButtonHandler(t *testing.T) {
	elevio.Init(elevServerAddr, numElevFloors)
	callsForSale := make(chan types.Call)
	buttonEvents := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(buttonEvents)
	StartButtonHandler(buttonEvents, callsForSale)
	for {
		call := <-callsForSale
		fmt.Printf("%+v\n", call)
	}
}
