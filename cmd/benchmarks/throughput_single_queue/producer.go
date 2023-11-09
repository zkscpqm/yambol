package main

import "yambol/pkg/queue"

func produceLoop(q *queue.Queue, val string, stop *bool) (total, successful int) {
	var err error
	for !*stop {
		_, err = q.Push(val)
		total++
		if err == nil {
			successful++
		}
	}
	q.Drain()
	return
}
