package log

import (
	"fmt"
	"runtime"
	"time"
)

type Formatter interface {
	Format(l Level, t time.Time, loggerName, s string, args ...interface{}) string
}

type DefaultFormatter struct {
	LevelNameMap   map[Level]string
	LevelColourMap map[Level]string
	TimeFormat     string
}

func NewDefaultLogFormatter() DefaultFormatter {
	return DefaultFormatter{
		LevelNameMap: map[Level]string{
			LevelDebug: "DEBUG",
			LevelInfo:  "INFO",
			LevelWarn:  "WARN",
			LevelError: "ERROR",
		},
		LevelColourMap: map[Level]string{
			-1:         "\033[0m", // -1 for the reset
			LevelError: "\033[31m",
			LevelInfo:  "\033[32m",
			LevelWarn:  "\033[33m",
			LevelDebug: "\033[97m",
		},
		TimeFormat: time.RFC3339, // The superior time format
	}
}

func (f DefaultFormatter) Format(l Level, t time.Time, loggerName, s string, args ...interface{}) string {
	llName, ok := f.LevelNameMap[l]
	if !ok {
		llName = "CUSTOM"
	}

	formatted := fmt.Sprintf(
		"[%s][%s][%s] %s",
		t.Format(f.TimeFormat),
		llName,
		loggerName,
		fmt.Sprintf(s, args...),
	)
	if runtime.GOOS == "windows" {
		return formatted
	}
	return f.Colour(l, formatted)
}

func (f DefaultFormatter) Colour(l Level, s string) string {
	col, ok := f.LevelColourMap[l]
	if !ok {
		return s
	}
	res, ok := f.LevelColourMap[-1]
	if !ok {
		res = "\033[0m"
	}
	return col + s + res
}
