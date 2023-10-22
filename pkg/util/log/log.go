package log

import (
	"fmt"
	"math"
	"path"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelOff = Level(^uint(0) >> 1)
)

var logLevelStringToLevelMap = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

func ParseLevel(level string) Level {
	if level == "" {
		return LevelInfo
	}
	l, ok := logLevelStringToLevelMap[level]
	if ok {
		return l
	}
	return LevelOff
}

func defaultLogFile() string {
	return path.Join(defaultLogDir, "yambol", "yambol.log")
}

type Logger struct {
	name     string
	level    Level
	handlers []Handler
}

func New(name string, level Level, handlers ...Handler) *Logger {
	for _, h := range handlers {
		h.SetLevel(level)
		h.SetName(name)
	}
	return &Logger{
		name:     name,
		level:    level,
		handlers: handlers,
	}
}

func NewStdoutOnly(name string) *Logger {
	return New(name, LevelInfo, NewDefaultStdioHandler())
}

func NewWithFile(name string, fileName string) (*Logger, error) {
	fh, err := NewDefaultFileHandler(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %v", fileName, err)
	}
	return New(name, LevelInfo, NewDefaultStdioHandler(), fh), nil
}

func NewDefault(name string) (*Logger, error) {
	return NewWithFile(name, defaultLogFile())
}

func (l *Logger) Output(code int, s string) error { // to allow for slack logger compatibility
	l.Info("%d %s", code, s)
	return nil
}

func (l *Logger) Log(level Level, format string, args ...interface{}) {
	level = Level(math.Min(float64(level), float64(LevelOff-1)))
	for _, h := range l.handlers {
		h.Log(level, format, args...)
	}
}
func (l *Logger) Debug(format string, args ...interface{}) {
	l.Log(LevelDebug, format, args...)
}
func (l *Logger) Info(format string, args ...interface{}) {
	l.Log(LevelInfo, format, args...)
}
func (l *Logger) Warn(format string, args ...interface{}) {
	l.Log(LevelWarn, format, args...)
}
func (l *Logger) Error(format string, args ...interface{}) {
	l.Log(LevelError, format, args...)
}
func (l *Logger) LogError(err error) {
	if err != nil {
		l.Error(err.Error())
	}
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
	for _, h := range l.handlers {
		h.SetLevel(level)
	}
}

func (l *Logger) GetLevel() Level {
	return l.level
}

func (l *Logger) SetName(name string) {
	l.name = name
	for _, h := range l.handlers {
		h.SetName(name)
	}
}

func (l *Logger) String() string {
	return fmt.Sprintf("Logger(%s)", l.Name())
}

func (l *Logger) Close() error {
	var lastErr error
	for _, h := range l.handlers {
		if err := h.Close(); err != nil {
			fmt.Printf("failed to close logger handle `%s`: %v", h.String(), err)
			lastErr = err
		}
	}
	return lastErr
}

func (l *Logger) NewFrom(name string) *Logger {
	newHandlers := make([]Handler, len(l.handlers))
	for i, h := range l.handlers {
		newHandler, err := h.NewFrom(name)
		if err != nil {
			l.Error("FATAL! Failed to copy logger `%s` -> `%s` handler `%s`...: %v", l.name, name, h.String(), err)
			return nil
		}
		newHandlers[i] = newHandler
	}
	return New(name, l.level, newHandlers...)
}

func (l *Logger) Off() {
	l.SetLevel(LevelOff)
}

func (l *Logger) Name() string {
	return l.name
}
