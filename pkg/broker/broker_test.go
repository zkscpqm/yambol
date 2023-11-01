package broker

import (
	"testing"
	"time"

	"yambol/config"
	"yambol/pkg/queue"
	"yambol/pkg/util/log"

	"github.com/stretchr/testify/assert"
)

const (
	testQueueDefaultMinLen       = int64(12)
	testQueueDefaultMaxLen       = int64(128)
	testQueueDefaultMaxSizeBytes = int64(1024 * 1024)
	testQueueDefaultTTL          = time.Minute
)

func setDefaults() {
	config.DisableAutoSave(true)
	SetDefaultMinLen(testQueueDefaultMinLen)
	SetDefaultMaxLen(testQueueDefaultMaxLen)
	SetDefaultMaxSizeBytes(testQueueDefaultMaxSizeBytes)
	SetDefaultTTL(testQueueDefaultTTL)
}

func testLogger() *log.Logger {
	return log.New("TEST", log.LevelOff)
}

func TestDefaults(t *testing.T) {

	config.DisableAutoSave(true)

	_defaultMinLen := GetDefaultMinLen()
	_defaultMaxLen := GetDefaultMaxLen()
	_defaultMaxSizeBytes := GetDefaultMaxSizeBytes()
	_defaultTTL := GetDefaultTTL()

	// test we won't ruin defaults with empty values
	SetDefaultMinLen(0)
	SetDefaultMaxLen(0)
	SetDefaultMaxSizeBytes(0)
	SetDefaultTTL(0)

	assert.Equal(t, _defaultMinLen, GetDefaultMinLen(), "failed to set new DefaultMinLen")
	assert.Equal(t, _defaultMaxLen, GetDefaultMaxLen(), "failed to set new DefaultMaxLen")
	assert.Equal(t, _defaultMaxSizeBytes, GetDefaultMaxSizeBytes(), "failed to set new DefaultMaxSizeBytes")
	assert.Equal(t, _defaultTTL, GetDefaultTTL(), "failed to set new DefaultTTL")

	setDefaults()

	assert.Equal(t, testQueueDefaultMinLen, GetDefaultMinLen(), "failed to set new DefaultMinLen")
	assert.Equal(t, testQueueDefaultMaxLen, GetDefaultMaxLen(), "failed to set new DefaultMaxLen")
	assert.Equal(t, testQueueDefaultMaxSizeBytes, GetDefaultMaxSizeBytes(), "failed to set new DefaultMaxSizeBytes")
	assert.Equal(t, testQueueDefaultTTL, GetDefaultTTL(), "failed to set new DefaultTTL")
}

func TestBrokerBasics(t *testing.T) {

	setDefaults()

	mb := New(testLogger())
	assert.Empty(t, mb.queues, "broker should have no queues by default")
	err := mb.RemoveQueue("test")
	assert.Error(t, err, "removed non existent queue")
	assert.False(t, mb.QueueExists("test"), "broker should have no queues by default")
	err = mb.AddDefaultQueue("test")
	assert.NoError(t, err, "failed to add test queue")
	assert.True(t, mb.QueueExists("test"), "broker should have test queue")
	assert.Len(t, mb.queues, 1, "broker should have 1 queue")
	assert.Len(t, mb.Stats(), 1, "broker should have 1 queue stats")
	assert.Len(t, mb.unsent, 1, "broker should have 1 unsent box")

	err = mb.AddDefaultQueue("test")
	assert.Error(t, err, "expected to fail to add duplicate queue name")

	err = mb.RemoveQueue("test")
	assert.NoError(t, err, "failed to remove test queue")

	assert.Len(t, mb.queues, 0, "broker should have 0 queues after deletion")
	assert.Len(t, mb.Stats(), 1, "broker queue stats for deleted queue should remain")
	assert.Len(t, mb.unsent, 1, "broker unsent box for deleted queue should remain")

}

func TestBrokerPublishConsume(t *testing.T) {

	setDefaults()

	mb := New(testLogger())

	_, err := mb.Consume("test")
	assert.Error(t, err, "expected to fail to consume from non existent queue")

	err = mb.AddDefaultQueue("test")
	assert.NoError(t, err, "failed to add test queue")

	_, err = mb.Consume("test")
	assert.ErrorAs(t, err, &queue.ErrQueueEmpty, "expected to fail to consume from empty queue")

	err = mb.Publish("my test message")
	assert.Error(t, err, "expected fail for: no queue was specified")

	err = mb.Publish("my test message", "nonexistentqueue")
	assert.Error(t, err, "expected fail for: an unknown queue was specified")

	err = mb.Publish("my test message", "test")
	assert.NoError(t, err, "failed to publish message")

	msg, err := mb.Consume("test")
	assert.NoError(t, err, "failed to consume message")
	assert.Equal(t, "my test message", msg, "failed to consume message")

	oneNs := time.Nanosecond
	err = mb.PublishWithTTL("fast disappearing", &oneNs, "test")
	assert.NoError(t, err, "failed to publish message")
	time.Sleep(time.Millisecond * 10)
	msg, err = mb.Consume("test")
	assert.ErrorAs(t, err, &queue.ErrQueueEmpty, "expected to fail to consume due to ttl reached")

}

func TestBrokerBroadcast(t *testing.T) {

	setDefaults()

	mb := New(testLogger())

	err := mb.AddDefaultQueue("test1")
	assert.NoError(t, err, "failed to add test1 queue")

	err = mb.AddDefaultQueue("test2")
	assert.NoError(t, err, "failed to add test2 queue")

	err = mb.Broadcast("my test message")
	assert.NoError(t, err, "broadcast failed")

	msg, err := mb.Consume("test1")
	assert.NoError(t, err, "consume failed for queue test1")
	assert.Equal(t, "my test message", msg, "consume failed")

	msg, err = mb.Consume("test2")
	assert.NoError(t, err, "consume failed for queue test2")
	assert.Equal(t, "my test message", msg, "consume failed")
}
