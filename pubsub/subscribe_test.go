package pubsub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sigtot/sanntid/utils"
	"net/http"
	"testing"
	"time"
)

type SubDude struct {
	WeekDay string `json:"WeekDay"`
}

func TestSubscriber(t *testing.T) {
	// Listen for published data
	receivedBufs, httpPort := StartSubscriber(41000, "sales")

	// Publish
	myDude := SubDude{WeekDay: "Wednesday"}
	myDudeJson, err := json.Marshal(myDude)
	if err != nil {
		t.Fatal("Could not marshal json")
	}
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d", httpPort), "application/json", bytes.NewBuffer(myDudeJson))
	if err != nil {
		fmt.Printf("Got response %x %x \n", resp.StatusCode, resp.Status)
	}
	err = resp.Body.Close()
	utils.OkOrPanic(err)

	select {
	case buf := <-receivedBufs:
		myReceivedDude := SubDude{}
		err = json.Unmarshal(buf, &myReceivedDude)
		if err != nil {
			t.Logf("Could not unmarshal json, %s\n", err.Error())
		}
		fmt.Printf("It is %s my dudes – subscriber\n", myReceivedDude.WeekDay)

	case <-time.After(300 * time.Millisecond):
		t.Fatal("Did not receive publish in 500ms")
	}
}
