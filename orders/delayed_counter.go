package orders

import (
	"time"
)

type DelayedCounter struct {
	count int
	Count chan int
	reset chan int
	stop  chan int
}

const (
	stateWaiting = iota
	stateCounting
)

func (dc *DelayedCounter) Start(delay time.Duration, countInterval time.Duration) {
	dc.Count = make(chan int)
	dc.reset = make(chan int)
	dc.stop = make(chan int)
	state := stateWaiting
	ticker := time.NewTicker(countInterval)
	delayC := time.After(delay)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-delayC:
				state = stateCounting
			case <-ticker.C:
				if state == stateCounting {
					dc.count++
				}
			case dc.Count <- dc.count:
			case <-dc.reset:
				delayC = time.After(delay)
				state = stateWaiting
				dc.count = 0
			case <-dc.stop:
				return
			}
		}
	}()
}

func (dc *DelayedCounter) Stop() {
	dc.stop <- 0
}

func (dc *DelayedCounter) Reset() {
	dc.reset <- 0
}
