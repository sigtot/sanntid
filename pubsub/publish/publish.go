package publish

import (
	"bytes"
	"fmt"
	"github.com/sigtot/sanntid/hotchan"
	"net"
	"net/http"
	"time"
)

const TTL = 5 * time.Second

// StartPublisher starts a publisher. It will listen for subscribers on the given discoveryPort.
// Items in the thingsToPublish channel will be published to all current subscribers.
func StartPublisher(discoveryPort int, thingsToPublish chan []byte) {
	discoveredIPs := make(chan string)
	go listenForSubscribers(discoveryPort, discoveredIPs)
	subHotChan := hotchan.HotChan{}
	subHotChan.Start()
	defer subHotChan.Stop()
	for {
		select {
		case ip := <-discoveredIPs:
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
		case thingToPublish := <-thingsToPublish:
			fanOutPublish(thingToPublish, subHotChan)
		}
	}
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
