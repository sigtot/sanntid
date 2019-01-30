package hotqueue

import "time"

type HotQueue struct {
	queue []Item
}

type Item struct {
	Val     interface{}
	TTL     time.Duration
	created time.Time
}

func (q *HotQueue) Push(item Item) {
	if item.created.IsZero() {
		item.created = time.Now()
	}
	q.queue = append(q.queue, item)
}

func (q *HotQueue) Pop() Item {
	for len(q.queue) > 0 {
		if time.Since(q.queue[0].created) > q.queue[0].TTL {
			q.queue = q.queue[1:]
		} else {
			var item Item
			item, q.queue = q.queue[0], q.queue[1:]
			return item
		}
	}
	return Item{Val: nil, TTL: 0 * time.Second}
}

func (q *HotQueue) Len() (len int) {
	for _, item := range q.queue {
		if time.Since(item.created) <= item.TTL {
			len++
		}
	}
	return len
}
