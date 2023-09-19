package queue

import "sync"

type Queue[_T ValueType] struct {
	mx      *sync.RWMutex
	minSize int
	maxSize int
	items   []item[_T]
	factory itemFactory[_T]
}

func New[_T ValueType](minSize, maxSize int) Queue[_T] {
	return Queue[_T]{
		mx:      &sync.RWMutex{},
		minSize: minSize,
		maxSize: maxSize,
		items:   make([]item[_T], 0, minSize),
		factory: newItemFactory[_T](),
	}
}

// Internal method that gets length without locking.
func (q *Queue[_T]) len() int {
	return len(q.items)
}

func (q *Queue[_T]) Len() int {
	q.mx.RLock()
	defer q.mx.RUnlock()
	return q.len()
}

func (q *Queue[_T]) PushBatch(values ..._T) ([]int, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len()+len(values) >= q.maxSize {
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

func (q *Queue[_T]) Push(value _T) (int, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len() == q.maxSize {
		return -1, ErrQueueFull
	}

	item_ := q.factory.newItem(value)
	q.items = append(q.items, item_)
	return item_.uid, nil
}

func (q *Queue[_T]) Pop() (*_T, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len() == 0 {
		return nil, ErrQueueEmpty
	}

	item_ := q.pop()
	q.factory.removeUid(item_.uid)
	return &item_.value, nil
}

func (q *Queue[_T]) Drain() []_T {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.len() == 0 {
		return []_T{}
	}

	values := make([]_T, q.len())
	for i, item_ := range q.items {
		values[i] = item_.value
	}
	q.clear()
	return values
}

func (q *Queue[_T]) pop() item[_T] {
	item_ := q.items[0]
	q.items = q.items[1:]

	// Shrink the underlying array when it is less than half full
	// chatgpt optimization magic idk but it works
	if cap(q.items) > q.minSize && len(q.items) < cap(q.items)/2 {
		newItems := make([]item[_T], len(q.items), len(q.items)*2)
		copy(newItems, q.items)
		q.items = newItems
	}

	return item_
}

func (q *Queue[_T]) clear() {
	q.items = make([]item[_T], 0, q.minSize)
	q.factory.clear()
}
