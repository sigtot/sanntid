package subscribe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type Dude struct {
	WeekDay string `json:"WeekDay"`
}

func TestSubscriber(t *testing.T) {
	// Listen for published data
	receivedBufs, httpPort := StartSubscriber(41000)

	// Publish
	myDude := Dude{WeekDay: "Wednesday"}
	myDudeJson, err := json.Marshal(myDude)
	if err != nil {
		t.Fatal("Could not marshal json")
	}
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d", httpPort), "application/json", bytes.NewBuffer(myDudeJson))
	if err != nil {
		fmt.Printf("Got response %x %x \n", resp.StatusCode, resp.Status)
	}
	err = resp.Body.Close()
	checkError(err)

	select {
	case buf := <-receivedBufs:
		myReceivedDude := Dude{}
		err = json.Unmarshal(buf, &myReceivedDude)
		if err != nil {
			t.Logf("Could not unmarshal json, %s\n", err.Error())
		}
		fmt.Printf("It is %s my dudes – subscriber\n", myReceivedDude.WeekDay)

	case <-time.After(300 * time.Millisecond):
		t.Fatal("Did not receive publish in 500ms")
	}
}
