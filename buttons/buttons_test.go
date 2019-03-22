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
	go StartButtonHandler(callsForSale)
	for {
		call := <-callsForSale
		fmt.Printf("%+v\n", call)
	}
}
