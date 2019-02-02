package hotchan

import (
	"testing"
	"time"
)

func TestHotChanStartStop(t *testing.T) {
	c := HotChan{}
	c.Start()
	c.Stop()
}

func TestHotChan(t *testing.T) {
	c := HotChan{}
	c.Start()
	defer c.Stop()

	c.In <- Item{Val: 1, TTL: 40 * time.Millisecond}
	c.In <- Item{Val: 2, TTL: 20 * time.Millisecond}
	c.In <- Item{Val: 3, TTL: 50 * time.Millisecond}

	time.Sleep(30 * time.Millisecond)
	one := <-c.Out
	three := <-c.Out
	if one.Val != 1 || three.Val != 3 {
		t.Errorf("Error: expected %d and %d but got %d and %d\n", 1, 3, one.Val, three.Val)
	}
}

// Items should decay outside of the hot channel as well as inside
func TestOutsideDecay(t *testing.T) {
	c := HotChan{}
	c.Start()
	defer c.Stop()

	c.In <- Item{Val: 42, TTL: 20 * time.Millisecond}

	time.Sleep(5 * time.Millisecond)
	item := <-c.Out

	// Wait for it to die outside the channel
	time.Sleep(20 * time.Millisecond)

	// Put it back in
	c.In <- item

	// Ensure that we're not able to retrieve it
L:
	for {
		select {
		case <-c.Out:
			t.Errorf("Error: Item not dead after its TTL ran out")
		case <-time.After(10 * time.Millisecond):
			break L
		}
	}

	// Ensure that out channel is empty
	if l := len(c.Out); l != 0 {
		t.Errorf("Error: Out channel should be empty, but it is length %d\n", l)
	}
}
