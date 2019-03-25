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
	"time"
)

const dbName = "orderwatcher.db"
const dbPerms = 0600
const dbTimeout = 300

const moduleName = "MAIN"

const defaultElevPort = 15657

func main() {
	rand.Seed(time.Now().UnixNano())

	var elevPort = flag.Int("port", defaultElevPort, "port for connecting to the elevator server")
	flag.Parse()

	log := logrus.New()
	utils.Log(log, moduleName, "Starting elevator")

	goalArrivals := make(chan types.Order)
	currentGoals := make(chan types.Order)
	floorArrivals := make(chan int)
	go elevio.PollFloorSensor(floorArrivals)
	elevator := elev.StartElevController(goalArrivals, currentGoals, floorArrivals, *elevPort)

	callsForSale := make(chan types.Call)
	buttonEvents := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(buttonEvents)
	buttons.StartButtonHandler(buttonEvents, callsForSale)

	indicators.StartIndicatorHandler()

	oh, newOrders := orders.StartOrderHandler(currentGoals, goalArrivals, elevator)

	buyer.StartBuying(oh, newOrders)

	go seller.StartSelling(callsForSale) //Go this

	orderWatcherDb, err := bolt.Open(dbName, dbPerms, &bolt.Options{Timeout: dbTimeout * time.Millisecond})
	if err != nil {
		panic(err)
	}
	quitOrderWatcher := make(chan int)
	orderWatcherQuitAck := orderwatcher.StartOrderWatcher(callsForSale, orderWatcherDb, quitOrderWatcher)

	quitDistributor := make(chan int)
	orderwatcher.StartDbDistributor(orderWatcherDb, dbName, quitDistributor)

	utils.Log(log, moduleName, "Successfully initialized all modules")

	sigInt := make(chan os.Signal, 1)
	signal.Notify(sigInt, os.Interrupt)
	<-sigInt
	signal.Stop(sigInt)
	utils.Log(log, moduleName, "Gracefully stopping all modules. Do ^C again to force")
	quitOrderWatcher <- 1
	<-orderWatcherQuitAck
	err = orderWatcherDb.Close()
	if err != nil {
		panic(err)
	}
}
