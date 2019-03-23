package publish

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/sigtot/sanntid/hotchan"
	"github.com/sigtot/sanntid/utils"
	"net"
	"net/http"
	"time"
)

const TTL = 5 * time.Second

// StartPublisher starts a publisher. It will listen for subscribers on the given discoveryPort.
// Items in the returned buffered channel will be published to all current subscribers.
func StartPublisher(discoveryPort int) chan []byte {
	thingsToPublish := make(chan []byte, 1024)
	discoveredIPs := make(chan string)
	go listenForSubscribers(discoveryPort, discoveredIPs)
	go func() {
		log := logrus.New()
		subHotChan := hotchan.HotChan{}
		subHotChan.Start()
		defer subHotChan.Stop()
		for {
			select {
			case ip := <-discoveredIPs:
				numSubsBefore := len(subHotChan.Out)
				subs := make(chan hotchan.Item, 1024)
				newSub := hotchan.Item{Val: ip, TTL: TTL}
				subs <- newSub
				for len(subHotChan.Out) > 0 {
					sub := <-subHotChan.Out
					if sub.Val != newSub.Val {
						subs <- sub
					}
				}
				for len(subs) > 0 {
					subHotChan.Insert(<-subs)
				}
				if len(subHotChan.Out) > numSubsBefore {
					utils.LogNewSub(log, "PUBLISHER", "Discovered new subscriber", newSub.Val.(string))
				}
			case thingToPublish := <-thingsToPublish:
				fanOutPublish(thingToPublish, subHotChan)
			}
		}
	}()
	return thingsToPublish
}

func listenForSubscribers(discoveryPort int, discoveredIPs chan string) {
	lAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", discoveryPort))
	checkError(err)

	conn, err := net.ListenUDP("udp", lAddr)
	checkError(err)
	defer func() {
		err := conn.Close()
		checkError(err)
	}()

	buf := make([]byte, 1024)
	for {
		_, addr, err := conn.ReadFromUDP(buf)
		checkError(err)
		discoveredIPs <- addr.String()
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func publish(addr string, body []byte) {
	resp, err := http.Post(fmt.Sprintf("http://%s", addr), "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Got response %d %s \n", resp.StatusCode, resp.Status)
	}
	err = resp.Body.Close()
	checkError(err)
}

// Publish thingToPublish to all subscribers in subHotChan.
// Must not be run concurrently.
func fanOutPublish(thingToPublish []byte, subHotChan hotchan.HotChan) {
	subs := make(chan hotchan.Item, 1024)
	for len(subHotChan.Out) > 0 {
		subs <- <-subHotChan.Out
	}
	for len(subs) > 0 {
		sub := <-subs
		go publish(sub.Val.(string), thingToPublish)
		subHotChan.Insert(sub)
	}
}
