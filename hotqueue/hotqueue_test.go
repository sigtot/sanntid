package hotqueue

import (
	"testing"
	"time"
)

func TestHotQueue(t *testing.T) {
	q := HotQueue{}
	q.Push(Item{Val: 1, TTL: 500 * time.Millisecond})
	q.Push(Item{Val: 2, TTL: 200 * time.Millisecond})
	q.Push(Item{Val: 3, TTL: 400 * time.Millisecond})

	time.Sleep(300 * time.Millisecond)
	if q.Len() != 2 {
		t.Errorf("Error: Length should be 2, but it is %d\n", q.Len())
	}
	if one, three := q.Pop(), q.Pop(); one.Val != 1 || three.Val != 3 {
		t.Errorf("%d, %d does not match %d, %d\n", one.Val, three.Val, 1, 3)
	}
}

func TestDoubleDip(t *testing.T) {
	q := HotQueue{}

	// Dip
	q.Push(Item{Val: 1, TTL: 300 * time.Millisecond})

	time.Sleep(200 * time.Millisecond)
	if q.Len() != 1 {
		t.Errorf("Error, length should be 1, but it is %d", q.Len())
	}

	item := q.Pop()

	// Double dip
	q.Push(item)

	time.Sleep(200 * time.Millisecond)
	if q.Len() != 0 {
		t.Errorf("Error, length should be 0, but it is %d", q.Len())
	}
}
