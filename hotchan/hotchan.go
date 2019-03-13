package hotchan

import (
	"sync"
	"time"
)

/*
HotChan is a channel with an In channel and an Out channel.
Items put into the hc.In channel will expire after their time to live (TTL) runs out, but will be available in FIFO
call from the hc.Out channel until then. The TTL countdown will run even if the items are not currently inside the
channel.
*/
type HotChan struct {
	In      chan Item
	Out     chan Item
	toPurge chan int
	quit    chan int
	status  statusMap
}

// Item to be held in the hot channel. Needs a Val and a TTL.
type Item struct {
	Val interface{}
	TTL time.Duration
	id  int
}

type statusMap struct {
	status map[int]chan int
	mu     sync.Mutex
}

// Start the HotQueue. Initializes channels and starts goroutines managing this hot channel
func (c *HotChan) Start() {
	c.In = make(chan Item, 1024)
	c.Out = make(chan Item, 1024)
	c.toPurge = make(chan int, 1024)
	c.quit = make(chan int)
	c.status = statusMap{status: make(map[int]chan int)}
	go c.manage()
}

// Stop sends a stop signal to the goroutine managing the hot channel
func (c *HotChan) Stop() {
	c.quit <- 0
}

// Manages incoming items
func (c *HotChan) manage() {
	var count int
	for {
		select {
		case item := <-c.In:
			c.status.mu.Lock()
			if item.id == 0 {
				count++
				item.id = count
				c.status.status[item.id] = make(chan int, 1)
				c.status.status[item.id] <- 1
				go c.doom(item)
			} else if len(c.status.status[item.id]) == 0 {
				break
			}
			c.status.mu.Unlock()
			c.Out <- item
		case killID := <-c.toPurge:
			spared := make(chan Item, 1024)
		L:
			for {
				select {
				case item := <-c.Out:
					if item.id != killID {
						spared <- item
					} else {
						continue L
					}
				default:
					break L
				}
			}
			for len(spared) > 0 {
				c.Out <- <-spared
			}
		case <-c.quit:
			return
		}
	}
}

// Kills an item after it's TTL runs out
func (c *HotChan) doom(item Item) {
	<-time.After(item.TTL)
	c.toPurge <- item.id

	c.status.mu.Lock()
	<-c.status.status[item.id] // Signal that this item is done
	c.status.mu.Unlock()
}
