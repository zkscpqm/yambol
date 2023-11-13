package queue

import (
	"testing"
	"time"

	"yambol/pkg/util"

	"github.com/stretchr/testify/assert"
)

const (
	testItemDefaultTTL = time.Millisecond * 300
)

func TestItemExpiry(t *testing.T) {
	i := item{
		uid:   1,
		value: "test",
		ts:    time.Now(),
		ttl:   testItemDefaultTTL,
		tiq:   new(time.Duration), // avoid nil pointer dereference
	}
	halfTTL := testItemDefaultTTL / 2

	*i.tiq = halfTTL // simulate being in queue. Better than sleeping...
	assert.False(t, i.Expired(), "item should not have expired yet")

	*i.tiq += halfTTL + 1 // simulate waiting util ttl expired
	assert.True(t, i.Expired(), "item should have expired")
}

func TestItemDequeue(t *testing.T) {
	i := item{
		uid:   1,
		value: "test",
		ts:    time.Now(),
		ttl:   testItemDefaultTTL,
	}
	i.dequeue()
	tiqPre := i.TimeInQueue()
	//atomic.StoreInt64(&tiqPre, i.TimeInQueue().Nanoseconds())
	time.Sleep(time.Millisecond * 1)
	tiqPost := i.TimeInQueue()
	assert.Equal(t, tiqPre, tiqPost, "time in queue changed after dequeue")
}

func TestItemFactory(t *testing.T) {
	f := newItemFactory(testItemDefaultTTL)
	i := f.newDefaultItem("test")
	assert.Contains(t, f.uidMap, i.uid, "could not find item UID in factory UID Map")
	assert.Equal(t, testItemDefaultTTL, i.ttl, "item ttl not set correctly by factory")
	f.removeUid(i.uid)
	assert.NotContains(t, f.uidMap, i.uid, "found item UID in factory UID Map after delete")

	n := 5
	for range util.Range(n) {
		f.newDefaultItem("")
	}
	assert.Len(t, f.uidMap, n, "incorrect uidMap length after populating")
	f.clear()
	assert.Len(t, f.uidMap, 0, "incorrect uidMap length after clear")
}
