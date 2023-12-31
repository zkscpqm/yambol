package queue

import (
	"sync"
	"time"
	"yambol/config"

	"yambol/pkg/telemetry"
)

type Queue struct {
	mx           *sync.RWMutex
	minLen       int64
	maxLen       int64
	maxSizeBytes int64
	items        []item
	factory      itemFactory
	stats        *telemetry.QueueStats
}

func New(cfg config.QueueConfig, stats *telemetry.QueueStats) *Queue {
	if cfg.MinLength <= 0 {
		cfg.MinLength = 1
	}
	return &Queue{
		stats:        stats,
		mx:           &sync.RWMutex{},
		minLen:       cfg.MinLength,
		maxLen:       cfg.MaxLength,
		maxSizeBytes: cfg.MaxSizeBytes,
		items:        make([]item, 0, cfg.MinLength),
		factory:      newItemFactory(cfg.TTLDuration()),
	}
}

func (q *Queue) len() int {
	return len(q.items)
}

func (q *Queue) len64() int64 {
	return int64(len(q.items))
}

func (q *Queue) Len() int {
	q.mx.RLock()
	defer q.mx.RUnlock()
	return q.len()
}

func (q *Queue) PushBatch(values ...string) ([]int, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if int64(q.len()+len(values)) >= q.maxLen {
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

func (q *Queue) PushWithTTL(value string, ttl *time.Duration) (int, error) {
	if ttl == nil {
		return q.Push(value)
	}
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len64() >= q.maxLen {
		return -1, ErrQueueFull
	}

	item_ := q.factory.newItem(value, *ttl)
	q.items = append(q.items, item_)
	return item_.uid, nil
}

func (q *Queue) Push(value string) (int, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len64() >= q.maxLen {
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
	if int64(cap(q.items)) > q.minLen && len(q.items) < cap(q.items)/2 {
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
