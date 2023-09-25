package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"strings"
	"time"
	"yambol/pkg/queue"
)

const (
	minMaxSize = 1024 * 1024 * 1024
	seconds    = 10
	oneByte    = "a"
)

func produceLoop(q *queue.Queue, val string, stop *bool) (total, successful int) {
	var err error
	for !*stop {
		_, err = q.Push(val)
		total++
		if err == nil {
			successful++
		}
	}
	return
}

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

func test(val string) {
	size := len(val)
	q := queue.New(minMaxSize, minMaxSize)

	stop := false
	prodTotal := 0
	prodSuccessful := 0
	consTotal := 0
	consSuccessful := 0

	go func() {
		prodTotal, prodSuccessful = produceLoop(&q, val, &stop)
	}()
	go func() {
		consTotal, consSuccessful = consumeLoop(&q, &stop)
	}()

	time.Sleep(time.Second * seconds)
	stop = true
	time.Sleep(time.Millisecond * 1)

	calculateVolume := func(n int) string {
		totalSize := float64(n * size)
		unit := "Bytes"

		if totalSize > math.Pow(1024, 4) { // If size is more than a Terabyte, show in TB
			totalSize = totalSize / math.Pow(1024, 4)
			unit = "TB"
		} else if totalSize > math.Pow(1024, 3) { // If size is more than a Gigabyte, show in GB
			totalSize = totalSize / math.Pow(1024, 3)
			unit = "GB"
		} else if totalSize > math.Pow(1024, 2) { // If size is more than a Megabyte, show in MB
			totalSize = totalSize / math.Pow(1024, 2)
			unit = "MB"
		} else if totalSize > 1024 { // If size is more than a Kilobyte, show in KB
			totalSize = totalSize / 1024
			unit = "KB"
		}

		return fmt.Sprintf("%.2f%s", totalSize, unit)
	}

	log.Printf("[%dB] Total Consume: %d (%s)", size, prodTotal, calculateVolume(prodTotal))
	log.Printf("[%dB] Total Produce: %d (%s)", size, consTotal, calculateVolume(consTotal))
	log.Printf("[%dB] Successful Consume: %d (%s)", size, prodSuccessful, calculateVolume(prodSuccessful))
	log.Printf("[%dB] Successful Produce: %d (%s)", size, consSuccessful, calculateVolume(consSuccessful))
	log.Printf("[%dB] Total Throughput: %d (%s/s)", size, consTotal/seconds, calculateVolume(consTotal/seconds))
	log.Printf("[%dB] Successful Throughput: %d (%s/s)", size, consSuccessful/seconds, calculateVolume(consSuccessful/seconds))
}

func main() {
	file, err := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	if err = pprof.StartCPUProfile(file); err != nil {
		fmt.Println(err)
		return
	}
	for _, sizes := range []int{8, 16, 512, 1024} {
		test(strings.Repeat(oneByte, sizes))
	}
	pprof.StopCPUProfile()
}
