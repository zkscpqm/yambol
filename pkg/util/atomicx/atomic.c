#include <stdatomic.h>
#include <stdint.h>
#include <stdio.h>

void maxSwap64(uint64_t* tgt, uint64_t val) {
    if(tgt == NULL) {
        return;
    }
    int64_t prev = *tgt;
    while (val > prev) {
        if (atomic_compare_exchange_strong(tgt, &prev, val)) break;
    }
}
