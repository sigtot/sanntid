package main

import (
	"flag"
	"github.com/sigtot/elevio"
	"github.com/sigtot/sanntid/buttons"
	"github.com/sigtot/sanntid/buyer"
	"github.com/sigtot/sanntid/elev"
	"github.com/sigtot/sanntid/indicators"
	"github.com/sigtot/sanntid/orders"
	"github.com/sigtot/sanntid/orderwatcher"
	"github.com/sigtot/sanntid/seller"
	"github.com/sigtot/sanntid/types"
	"github.com/sigtot/sanntid/utils"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"
)

const dbName = "orderwatcher.db"
const dbPerms = 0600
const dbTimeout = 300

const moduleName = "MAIN"

const defaultElevPort = 15657

func main() {
	rand.Seed(time.Now().UnixNano())

	// Read elevator server port flag
	var elevPort = flag.Int("port", defaultElevPort, "port for connecting to the elevator server")
	flag.Parse()

	log := logrus.New()
	utils.Log(log, moduleName, "Starting elevator")

	var wg sync.WaitGroup

	goalArrivals := make(chan types.Order)
	currentGoals := make(chan types.Order)
	floorArrivals := make(chan int)
	quitElev := make(chan int)
	go elevio.PollFloorSensor(floorArrivals)
	elevator := elev.StartElevController(goalArrivals, currentGoals, floorArrivals, *elevPort, quitElev, &wg)

	callsForSale := make(chan types.Call)
	buttonEvents := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(buttonEvents)
	buttons.StartButtonHandler(buttonEvents, callsForSale)

	quitIndicators := make(chan int)
	indicators.StartIndicatorHandler(quitIndicators, &wg)

	oh, newOrders := orders.StartOrderHandler(currentGoals, goalArrivals, elevator)

	buyer.StartBuying(oh, newOrders)

	seller.StartSelling(callsForSale)

	orderWatcherDb, err := bolt.Open(dbName, dbPerms, &bolt.Options{Timeout: dbTimeout * time.Millisecond})
	if err != nil {
		panic(err)
	}
	quitOrderWatcher := make(chan int)
	orderwatcher.StartOrderWatcher(callsForSale, orderWatcherDb, quitOrderWatcher, &wg)

	quitDistributor := make(chan int)
	orderwatcher.StartDbDistributor(orderWatcherDb, dbName, quitDistributor)

	utils.Log(log, moduleName, "Successfully initialized all modules")

	sigInt := make(chan os.Signal, 1)
	signal.Notify(sigInt, os.Interrupt)
	<-sigInt
	signal.Stop(sigInt) // Stop trapping interrupt signal to give it back its usual behavior

	utils.Log(log, moduleName, "Gracefully stopping all modules. Do ^C again to force")
	quitElev <- 0
	quitIndicators <- 0
	quitOrderWatcher <- 0
	err = orderWatcherDb.Close()
	if err != nil {
		panic(err)
	}
	wg.Wait()
	utils.Log(log, moduleName, "Stopped elevator")
}
