package queue

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ValueType interface {
	string | int | float64
}

type item[_t ValueType] struct {
	uid   int
	value _t
}

type itemFactory[_t ValueType] struct {
	uidMap map[int]struct{}
	mx     *sync.RWMutex
}

func newItemFactory[_t ValueType]() itemFactory[_t] {
	return itemFactory[_t]{
		uidMap: make(map[int]struct{}),
		mx:     &sync.RWMutex{},
	}
}

func (f *itemFactory[_t]) generateUid() int {
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

func (f *itemFactory[_t]) removeUid(uid int) {
	f.mx.Lock()
	defer f.mx.Unlock()
	delete(f.uidMap, uid)
}

func (f *itemFactory[_t]) clear() {
	f.mx.Lock()
	defer f.mx.Unlock()
	f.uidMap = make(map[int]struct{})
}

func (f *itemFactory[_t]) newItem(val _t) item[_t] {
	return item[_t]{
		uid:   f.generateUid(),
		value: val,
	}
}
