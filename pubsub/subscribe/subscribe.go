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

const aliveSignalInterval  = 300

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
func StartSubscriber(discoveryPort int) (receivedBufs chan []byte, httpPort int) {
	receivedBufs = make(chan []byte, 1024)
	httpPort = findAvailPort()
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			subHandler(w, r, receivedBufs)
		})
		if err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux); err != nil {
			panic(err)
		}
	}()
	time.Sleep(500 * time.Millisecond) // Wait for server to start
	go sendAliveSignal(discoveryPort, httpPort)
	return receivedBufs, httpPort
}

func subHandler(w http.ResponseWriter, r *http.Request, receivedBufs chan []byte) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Could not read request body: %s", err.Error())
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
	}

	receivedBufs <- buf
	w.WriteHeader(http.StatusOK)
}

func sendAliveSignal(discoveryPort int, publishPort int) {
	sAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", discoveryPort))
	okOrPanic(err)
	lAddr, err := net.ResolveUDPAddr("udp",fmt.Sprintf("localhost:%d", discoveryPort))
	okOrPanic(err)
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", publishPort))
	okOrPanic(err)
	defer func() {
		err := conn.Close()
		okOrPanic(err)
	}()

	for {
		_, err = conn.WriteTo([]byte("sup"), sAddr)
		if err != nil {
			if strings.Contains(err.Error(), "network is unreachable"){
				_, err = conn.WriteTo([]byte("sup"), lAddr)
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
