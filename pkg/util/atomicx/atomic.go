package atomicx

/*
#include "atomic.h"
*/
import "C"

func MaxSwap64(tgt *uint64, val uint64) {
	if tgt == nil {
		return
	}
	C.maxSwap64((*C.uint64_t)(tgt), C.uint64_t(val))
}
