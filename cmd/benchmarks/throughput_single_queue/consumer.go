package main

import "yambol/pkg/queue"

func consumeLoop(q *queue.Queue, stop *bool) (total, successful int) {
	for !*stop {
		_, err := q.Pop()
		total++
		if err == nil {
			successful++
		}
	}
	return
}
