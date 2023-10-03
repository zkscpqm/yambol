package atomicx

import "sync/atomic"

func MaxSwap64(tgt *int64, val int64) {
	if tgt == nil {
		return
	}
	for {
		old := atomic.LoadInt64(tgt)
		if val <= old {
			break
		}
		if atomic.CompareAndSwapInt64(tgt, old, val) {
			break // The value was successfully swapped.
		}
	}
}
