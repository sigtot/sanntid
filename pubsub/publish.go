package pubsub

import (
	"bytes"
	"fmt"
	"github.com/sigtot/sanntid/hotchan"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"time"
)

const ttl = 5 * time.Second

const moduleName = "PUBLISHER"
const logString = "%-15s%s"

type subscriber struct {
	IP    string
	Topic string
}

// StartPublisher starts a publisher. It will listen for subscribers on the given discoveryPort.
// Items in the returned buffered channel will be published to all current subscribers.
func StartPublisher(discoveryPort int) chan []byte {
	thingsToPublish := make(chan []byte, 1024)
	discoveredSubs := make(chan subscriber)
	go listenForSubscribers(discoveryPort, discoveredSubs)
	go func() {
		log := logrus.New()
		subHotChan := hotchan.HotChan{}
		subHotChan.Start()
		defer subHotChan.Stop()
		for {
			select {
			case sub := <-discoveredSubs:
				subIP := sub.IP
				numSubsBefore := len(subHotChan.Out)
				subs := make(chan hotchan.Item, 1024)
				newSubI := hotchan.Item{Val: subIP, TTL: ttl}
				subs <- newSubI
				for len(subHotChan.Out) > 0 {
					subI := <-subHotChan.Out
					if subI.Val != newSubI.Val {
						subs <- subI
					}
				}
				for len(subs) > 0 {
					subHotChan.Insert(<-subs)
				}
				if len(subHotChan.Out) > numSubsBefore {
					logNewSub(log, moduleName, "Discovered new subscriber", sub)
				}
			case thingToPublish := <-thingsToPublish:
				fanOutPublish(thingToPublish, subHotChan)
			}
		}
	}()
	return thingsToPublish
}

// listenForSubscribers listens for heartbeat signals for subscribers, on a designated discovery-port.
// The discovery port are preassigned to a topic. All active subscribers are passed to the discoveredSubs channel.
func listenForSubscribers(discoveryPort int, discoveredSubs chan subscriber) {
	lAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", discoveryPort))
	utils.OkOrPanic(err)

	conn, err := net.ListenUDP("udp", lAddr)
	utils.OkOrPanic(err)
	defer func() {
		err := conn.Close()
		utils.OkOrPanic(err)
	}()

	buf := make([]byte, 1024)
	for {
		_, addr, err := conn.ReadFromUDP(buf)
		utils.OkOrPanic(err)
		topic := strings.TrimRight(string(buf), "\x00") // Trim away zero values from buf when converting to string
		sub := subscriber{IP: addr.String(), Topic: topic}
		discoveredSubs <- sub
	}
}

// publish posts the body of a messages to the specified address.
func publish(addr string, body []byte) {
	resp, err := http.Post(fmt.Sprintf("http://%s", addr), "application/json", bytes.NewBuffer(body))
	if err != nil {
		func() {
			errStrings := []string{
				"connection refused",
				"network is unreachable",
				"i/o timeout",
				"connection reset by peer",
			}
			for _, errStr := range errStrings {
				if strings.Contains(err.Error(), errStr) {
					logrus.WithFields(logrus.Fields{
						"IP": addr,
					}).Warnf(logString, moduleName, "Could not publish")
					return
				}
			}
			panic(err)
		}()
	}
	if resp != nil {
		err = resp.Body.Close()
		utils.OkOrPanic(err)
	}
}

// Publish thingToPublish to all subscribers in subHotChan.
// Must not be run concurrently for the same publisher.
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

func logNewSub(log *logrus.Logger, moduleName string, info string, sub subscriber) {
	log.WithFields(logrus.Fields{
		"IP":    sub.IP,
		"topic": sub.Topic,
	}).Infof(logString, moduleName, info)
}
