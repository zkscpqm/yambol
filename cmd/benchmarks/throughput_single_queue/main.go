package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"strings"
	"time"
	"unsafe"
	"yambol/pkg/queue"
)

const (
	minMaxSize  = 1024 * 1024 * 1024
	seconds     = 10
	smallString = "abcdefghijklmnop" // 16 bytes
	largeString = "a"                // Will be multiplied to 1024 bytes
)

func produceLoop[_T queue.ValueType](q *queue.Queue[_T], val _T, stop *bool) (total, successful int) {
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

func consumeLoop[_T queue.ValueType](q *queue.Queue[_T], stop *bool) (total, successful int) {
	noop := func(_T) {} // To help dereferencing
	for !*stop {
		pVal, err := q.Pop()
		if pVal != nil {
			noop(*pVal)
		}
		total++
		if err == nil {
			successful++
		}
	}
	return
}

func test[_T queue.ValueType](val _T, label string, size int) {
	q := queue.New[_T](minMaxSize, minMaxSize)

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

	log.Printf("[%s-%d] Total Consume: %d (%s)", label, size, prodTotal, calculateVolume(prodTotal))
	log.Printf("[%s-%d] Total Produce: %d (%s)", label, size, consTotal, calculateVolume(consTotal))
	log.Printf("[%s-%d] Successful Consume: %d (%s)", label, size, prodSuccessful, calculateVolume(prodSuccessful))
	log.Printf("[%s-%d] Successful Produce: %d (%s)", label, size, consSuccessful, calculateVolume(consSuccessful))
	log.Printf("[%s-%d] Total Throughput: %d (%s/s)", label, size, consTotal/seconds, calculateVolume(consTotal/seconds))
	log.Printf("[%s-%d] Successful Throughput: %d (%s/s)", label, size, consSuccessful/seconds, calculateVolume(consSuccessful/seconds))
}

func testInt() {
	v := int(0)
	test[int](v, "INT", int(unsafe.Sizeof(v)))
}

func testFloat64() {
	v := float64(0.0)
	test[float64](v, "FLOAT64", int(unsafe.Sizeof(v)))
}

func testString(content string) {
	test[string](content, "STRING", len(content))
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
	testInt()
	testFloat64()
	testString(smallString)
	testString(strings.Repeat(largeString, 1024))
	pprof.StopCPUProfile()
}
