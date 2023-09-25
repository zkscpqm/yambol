package queue

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

const (
	TestMinSize = 10
	TestMaxSize = 10000
)

func setUp() Queue {
	return New(TestMinSize, TestMaxSize)
}

func stringRange(n int) (rv []string) {
	rv = make([]string, n)
	for i := 0; i < n; i++ {
		rv[i] = strconv.Itoa(i)
	}
	return
}

func TestQueueBasicOps(t *testing.T) {
	q := setUp()
	var err error
	for i := 0; i < TestMaxSize; i++ {
		_, err = q.Push(strconv.Itoa(i))
		assert.NoError(t, err, "failed to push", i)
	}
	assert.Equal(t, q.Len(), TestMaxSize, "mismatched number of items")
	_, err = q.Push("oob")
	assert.Error(t, err, "managed to push after max size has been reached")
	var val string
	for i := 0; i < TestMaxSize; i++ {
		val, err = q.Pop()
		assert.NoError(t, err, "failed to pop", i)
		assert.NotNil(t, val, "empty value popped", i)
		assert.Equal(t, strconv.Itoa(i), val, "mismatched value popped", i, val)
	}
	assert.Empty(t, q.items, "queue not empty")
	_, err = q.Pop()
	assert.Error(t, err, "popped from empty queue")
}

func TestQueueBulkOps(t *testing.T) {
	q := setUp()
	var err error
	_, err = q.PushBatch(stringRange(TestMaxSize - 9)...)
	assert.NoError(t, err, "failed to push batch")
	assert.Equal(t, q.Len(), TestMaxSize-9, "mismatched number of items")
	_, err = q.PushBatch(stringRange(10)...)
	assert.Error(t, err, "managed to push over max size")
	vals := q.Drain()
	assert.Equal(t, len(vals), TestMaxSize-9, "mismatched number of drained items")
}
