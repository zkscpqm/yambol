package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"yambol/pkg/queue"
	"yambol/pkg/telemetry"
	"yambol/pkg/util/log"
)

const (
	minMaxLen = 1024 * 1024
	maxSize   = 1024 * 1024 * 32
	seconds   = 30
	oneByte   = "a"
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

func test(val string, logger *log.Logger) {
	size := len(val)

	q := queue.New(minMaxLen, minMaxLen, maxSize, 0, &telemetry.QueueStats{})

	stop := false
	prodTotal := 0
	prodSuccessful := 0
	consTotal := 0
	consSuccessful := 0

	go func() {
		prodTotal, prodSuccessful = produceLoop(q, val, &stop)
	}()
	go func() {
		consTotal, consSuccessful = consumeLoop(q, &stop)
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

	logger.Info("[%dB] Total Consume: %d (%s)", size, prodTotal, calculateVolume(prodTotal))
	logger.Info("[%dB] Total Produce: %d (%s)", size, consTotal, calculateVolume(consTotal))
	logger.Info("[%dB] Successful Consume: %d (%s)", size, prodSuccessful, calculateVolume(prodSuccessful))
	logger.Info("[%dB] Successful Produce: %d (%s)", size, consSuccessful, calculateVolume(consSuccessful))
	logger.Info("[%dB] Total Throughput: %d (%s/s)", size, consTotal/seconds, calculateVolume(consTotal/seconds))
	logger.Info("[%dB] Successful Throughput: %d (%s/s)", size, consSuccessful/seconds, calculateVolume(consSuccessful/seconds))
}

func main() {

	fh, err := log.NewDefaultFileHandler("./.logs/logging_overhead_WriteSync.log")
	logger := log.New("log_overhead", log.LevelDebug, fh, log.NewDefaultStdioHandler())
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer file.Close()
	for _, size := range []int{16, 256, 512, 1024, 2048} {
		test(strings.Repeat(oneByte, size), logger)
	}
}
