package publish

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type Dude struct {
	WeekDay string `json:"WeekDay"`
}

func TestPublish(t *testing.T) {
	quit := make(chan int)
	// Listen for published data
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			buf, err := ioutil.ReadAll(r.Body)

			myReceivedDude := Dude{}
			err = json.Unmarshal(buf, &myReceivedDude)
			if err != nil {
				t.Logf("Could not unmarshal json, %s\n", err.Error())
				http.Error(w, "500 internal server error", http.StatusInternalServerError)
			}

			fmt.Printf("It is %s my dudes\n", myReceivedDude.WeekDay)
			_, err = w.Write([]byte("My dudes"))
			if err != nil {
				t.Fatal("Could not write to body lol")
			}
			quit <- 0
		})
		if err := http.ListenAndServe(":51000", nil); err != nil {
			panic(err)
		}
	}()

	// Publish
	myDude := Dude{WeekDay: "Wednesday"}
	myDudeJson, err := json.Marshal(myDude)
	if err != nil {
		t.Fatal("Could not marshal json")
	}
	go publish("http://localhost:51000", myDudeJson)

	select {
	case <-quit:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Did not receive publish in 500ms")
	}
}
