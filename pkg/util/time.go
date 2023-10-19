package util

import (
	"time"
)

func Seconds(n int64) time.Duration {
	return time.Duration(n) * time.Second
}

func LittleLongerThan(d time.Duration) time.Duration {
	return d + time.Microsecond*50
}

func DurationAlmostEqual(expected, actual time.Duration, maxOffset time.Duration) bool {
	return (expected - actual).Abs() <= maxOffset
}
