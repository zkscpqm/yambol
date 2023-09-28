package atomicx

import "sync/atomic"

func MaxSwap64(tgt *uint64, val uint64) {
	if tgt == nil {
		return
	}
	for {
		old := atomic.LoadUint64(tgt)
		if val <= old {
			break
		}
		if atomic.CompareAndSwapUint64(tgt, old, val) {
			break // The value was successfully swapped.
		}
	}
}
