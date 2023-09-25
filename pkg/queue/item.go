package queue

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type item struct {
	uid   int
	value string
	ts    time.Time
	ttl   time.Duration
}

func (i *item) String() string {
	return i.value
}

func (i *item) Expired() bool {
	if i.ttl == 0 {
		return false
	}
	return time.Now().After(i.ts.Add(i.ttl))
}

type itemFactory struct {
	uidMap map[int]struct{}
	mx     *sync.RWMutex
}

func newItemFactory() itemFactory {
	return itemFactory{
		uidMap: make(map[int]struct{}),
		mx:     &sync.RWMutex{},
	}
}

func (f *itemFactory) generateUid() int {
	f.mx.Lock()
	defer f.mx.Unlock()
	for {
		uid := rand.Int()
		if _, ok := f.uidMap[uid]; !ok {
			f.uidMap[uid] = struct{}{}
			return uid
		}
	}
}

func (f *itemFactory) removeUid(uid int) {
	f.mx.Lock()
	defer f.mx.Unlock()
	delete(f.uidMap, uid)
}

func (f *itemFactory) clear() {
	f.mx.Lock()
	defer f.mx.Unlock()
	f.uidMap = make(map[int]struct{})
}

func (f *itemFactory) newItem(val string) item {
	return item{
		uid:   f.generateUid(),
		value: val,
		ts:    time.Now(),
	}
}
