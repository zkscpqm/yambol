package util

import "time"

func Seconds(n int) time.Duration {
	return time.Duration(n) * time.Second
}

func LittleLongerThan(d time.Duration) time.Duration {
	return d + time.Microsecond*50
}
