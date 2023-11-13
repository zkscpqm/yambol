package main

import (
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"strings"
	"time"
	"yambol/config"

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

func calculateVolume(n, size int) string {
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

func test(val string, logger *log.Logger) {
	size := len(val)

	q := queue.New(config.QueueConfig{
		MaxSizeBytes: maxSize,
		MinLength:    minMaxLen,
		MaxLength:    minMaxLen,
	}, &telemetry.QueueStats{})

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
	time.Sleep(time.Second * 1)

	logger.Info("[%dB] Total Consume: %d (%s)", size, prodTotal, calculateVolume(prodTotal, size))
	logger.Info("[%dB] Total Produce: %d (%s)", size, consTotal, calculateVolume(consTotal, size))
	logger.Info("[%dB] Successful Consume: %d (%s)", size, prodSuccessful, calculateVolume(prodSuccessful, size))
	logger.Info("[%dB] Successful Produce: %d (%s)", size, consSuccessful, calculateVolume(consSuccessful, size))
	logger.Info("[%dB] Total Throughput: %d (%s/s)", size, consTotal/seconds, calculateVolume(consTotal/seconds, size))
	logger.Info("[%dB] Successful Throughput: %d (%s/s)", size, consSuccessful/seconds, calculateVolume(consSuccessful/seconds, size))
}

func main() {

	fh, err := log.NewDefaultFileHandler("./.logs/throughput.log")
	logger := log.New("throughput", log.LevelDebug, fh, log.NewDefaultStdioHandler())
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile("cpu.prof", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer file.Close()
	if err = pprof.StartCPUProfile(file); err != nil {
		logger.Error(err.Error())
		return
	}
	for _, size := range []int{16, 256, 512, 1024, 2048, 9999, 99999} {
		test(strings.Repeat(oneByte, size), logger)
	}
	pprof.StopCPUProfile()
}
