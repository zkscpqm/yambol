package queue

import (
	"sync"
)

type Queue struct {
	mx           *sync.RWMutex
	minLen       int
	maxLen       int
	maxSizeBytes int
	items        []item
	factory      itemFactory
}

func New(minLen, maxLen, maxSizeBytes int) *Queue {
	if minLen <= 0 {
		minLen = 1
	}
	return &Queue{
		mx:           &sync.RWMutex{},
		minLen:       minLen,
		maxLen:       maxLen,
		maxSizeBytes: maxSizeBytes,
		items:        make([]item, 0, minLen),
		factory:      newItemFactory(),
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
		item_ := q.factory.newItem(value)
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

	item_ := q.factory.newItem(value)
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
	q.factory.removeUid(item_.uid)
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

	values := make([]string, q.len())
	for i, item_ := range q.items {
		values[i] = item_.value
	}
	q.clear()
	return values
}

func (q *Queue) pop() item {
	item_ := q.items[0]
	q.items = q.items[1:]

	// Shrink the underlying array when it is less than half full
	// chatgpt optimization magic idk but it works
	if cap(q.items) > q.minLen && len(q.items) < cap(q.items)/2 {
		newItems := make([]item, len(q.items), len(q.items)*2)
		copy(newItems, q.items)
		q.items = newItems
	}

	return item_
}

func (q *Queue) clear() {
	q.items = make([]item, 0, q.minLen)
	q.factory.clear()
}
