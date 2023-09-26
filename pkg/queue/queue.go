package queue

import (
	"sync"
	"time"
	"yambol/pkg/telemetry"
)

type Queue struct {
	mx           *sync.RWMutex
	minLen       int
	maxLen       int
	maxSizeBytes int
	items        []item
	factory      itemFactory
	stats        *telemetry.QueueStats
}

func New(minLen, maxLen, maxSizeBytes int, defaultTTL time.Duration, stats *telemetry.QueueStats) *Queue {
	if minLen <= 0 {
		minLen = 1
	}
	return &Queue{
		stats:        stats,
		mx:           &sync.RWMutex{},
		minLen:       minLen,
		maxLen:       maxLen,
		maxSizeBytes: maxSizeBytes,
		items:        make([]item, 0, minLen),
		factory:      newItemFactory(defaultTTL),
	}
}

func (q *Queue) len() int {
	return len(q.items)
}

func (q *Queue) Len() int {
	q.mx.RLock()
	defer q.mx.RUnlock()
	return q.len()
}

func (q *Queue) PushBatch(values ...string) ([]int, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len()+len(values) >= q.maxLen {
		return nil, ErrQueueFull
	}

	uids := make([]int, len(values))
	for i, value := range values {
		item_ := q.factory.newDefaultItem(value)
		q.items = append(q.items, item_)
		uids[i] = item_.uid
	}
	return uids, nil
}

func (q *Queue) Push(value string) (int, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len() == q.maxLen {
		return -1, ErrQueueFull
	}

	item_ := q.factory.newDefaultItem(value)
	q.items = append(q.items, item_)
	return item_.uid, nil
}

func (q *Queue) Pop() (string, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len() == 0 {
		return "", ErrQueueEmpty
	}
	item_ := q.pop()
	for item_.Expired() {
		q.factory.removeUid(item_.uid)
		q.stats.Drop(item_.TimeInQueue())
		if q.len() == 0 {
			return "", ErrQueueEmpty
		}
		item_ = q.pop()
	}
	q.stats.Process(item_.TimeInQueue())
	return item_.value, nil
}

func (q *Queue) peek() *item {
	q.mx.RLock()
	defer q.mx.RUnlock()
	if q.len() == 0 {
		return nil
	}
	return &q.items[0]
}

func (q *Queue) Drain() []string {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len() == 0 {
		return []string{}
	}

	values := make([]string, 0, len(q.items))
	for _, item_ := range q.items {
		item_.dequeue()
		if item_.Expired() {
			q.stats.Drop(item_.TimeInQueue())
		} else {
			q.stats.Process(item_.TimeInQueue())
			values = append(values, item_.value)
		}
	}
	q.clear()
	return values
}

func (q *Queue) resize() {
	if cap(q.items) > q.minLen && len(q.items) < cap(q.items)/2 {
		newItems := make([]item, len(q.items), len(q.items)*2)
		copy(newItems, q.items)
		q.items = newItems
	}
}

func (q *Queue) pop() item {
	item_ := q.items[0]
	item_.dequeue()
	if q.len() != 0 {
		q.items = q.items[1:]
		q.resize()
	}
	return item_
}

func (q *Queue) clear() {
	q.items = make([]item, 0, q.minLen)
	q.factory.clear()
}
