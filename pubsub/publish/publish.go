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
			subItems := make(chan hotchan.Item, 1024)
			newSubItem := hotchan.Item{Val: ip, TTL: TTL}
			subItems <- newSubItem
			for len(subHotChan.Out) > 0 {
				subItem := <-subHotChan.Out
				if subItem.Val != newSubItem.Val {
					subItems <- subItem
				}
			}
			for len(subItems) > 0 {
				subHotChan.In <- <-subItems
			}
		case thingToPublish := <-thingsToPublish:
			subItems := make(chan hotchan.Item, 1024)
			for len(subHotChan.Out) > 0 {
				subItems <- <-subHotChan.Out
			}
			for len(subItems) > 0 {
				subItem := <-subItems
				go publish(subItem.Val.(string), thingToPublish)
				subHotChan.In <- subItem
			}
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
	resp, err := http.Post(addr, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Got response %x %x \n", resp.StatusCode, resp.Status)
	}
	err = resp.Body.Close()
	checkError(err)
}
