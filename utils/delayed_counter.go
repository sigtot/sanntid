package utils

import (
	"time"
)

// DelayedCounter holds a count which can be read from the Count channel.
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

// Start starts a delayed counter. After a time interval specified by the delay argument,
// it will start counting with a period specified by the countInterval argument.
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

// Reset resets a DelayedCounter such that the count is set to zero and the delay is reset.
func (dc *DelayedCounter) Reset() {
	dc.reset <- 0
}

// Stop turns off a delayedCounter.
func (dc *DelayedCounter) Stop() {
	dc.stop <- 0
}
