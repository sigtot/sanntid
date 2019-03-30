/*
Package subscribe contains functionality for sending heartbeat signals,
and setting up http-servers for communications between nodes.
*/
package subscribe

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const aliveSignalInterval = 300

// findAvailPort searches for an available port for the tcp connection to use.
// The ports are randomly selected in a range fro port 10000 to 50000.
func findAvailPort() (port int) {
	for {
		port = rand.Intn(40000) + 10000
		conn, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(port)))
		if err != nil {
			if !strings.Contains(err.Error(), "address already in use") {
				panic(err)
			}
		}
		if conn != nil {
			err = conn.Close()
			if err != nil {
				panic(err)
			}
			break
		}
	}
	return port
}

// StartSubscriber starts a subscriber with a given discoveryPort and publishPort.
// Received items are made available in the returned channel.
// The returned httpPort is the port of the subscriber's http server
func StartSubscriber(discoveryPort int, topic string) (receivedBuffs chan []byte, httpPort int) {
	receivedBuffs = make(chan []byte, 1024)
	httpPort = findAvailPort()
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			subHandler(w, r, receivedBuffs)
		})
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux); err != nil {
			panic(err)
		}
	}()
	time.Sleep(500 * time.Millisecond) // Wait for server to start
	go sendAliveSignal(discoveryPort, httpPort, topic)
	return receivedBuffs, httpPort
}

// subHandler reads the http-requests, checks for errors,
// and passes the body of the requests to the receiver channel.
func subHandler(w http.ResponseWriter, r *http.Request, receivedBuffs chan []byte) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Could not read request body: %s", err.Error())
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
	}

	receivedBuffs <- buf
	w.WriteHeader(http.StatusOK)
}

// sendAliveSignal sends heartbeat signals on the subnet with a predetermined port.
// The port corresponds to the topic of the subscriber.
func sendAliveSignal(discoveryPort int, publishPort int, topic string) {
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", discoveryPort))
	okOrPanic(err)
	lAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", discoveryPort))
	okOrPanic(err)
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", publishPort))
	okOrPanic(err)
	defer func() {
		err := conn.Close()
		okOrPanic(err)
	}()

	for {
		_, err = conn.WriteTo([]byte(topic), sAddr)
		if err != nil {
			if strings.Contains(err.Error(), "network is unreachable") {
				_, err = conn.WriteTo([]byte(topic), lAddr)
				okOrPanic(err)
			} else {
				panic(err)
			}
		}
		time.Sleep(aliveSignalInterval * time.Millisecond)
	}
}

func okOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
