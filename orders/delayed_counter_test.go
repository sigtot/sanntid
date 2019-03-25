package orders

import (
	"testing"
	"time"
)

func TestDelayedCounter(t *testing.T) {
	var dc DelayedCounter
	dc.Start(300*time.Millisecond, 500*time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	if c := <-dc.Count; c != 0 {
		t.Fatalf("Timeout not reached, so count should be 0, but it is %d\n", c)
	}
	time.Sleep(300 * time.Millisecond)
	if c := <-dc.Count; c > 2 {
		t.Fatalf("Timeout has been reached, so we should have a large than 2 count, but it is %d\n", c)
	}

	dc.Reset()
	if c := <-dc.Count; c != 0 {
		t.Fatalf("Count should be 0 after reset, but it is %d\n", c)
	}

	dc.Stop()
}
