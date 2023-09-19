package queue

import "fmt"

var ErrQueueFull = fmt.Errorf("queue is full")
var ErrQueueEmpty = fmt.Errorf("queue is empty")
