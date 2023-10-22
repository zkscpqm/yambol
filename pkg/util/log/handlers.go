package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Handler interface {
	io.Closer
	fmt.Stringer
	Log(level Level, format string, args ...interface{})
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	SetLevel(level Level)
	SetName(name string)
	NewFrom(name string) (Handler, error)
	Close() error
}

type StdioHandler struct {
	loggerName string
	level      Level
	formatter  Formatter
	mx         *sync.Mutex
}

func NewStdioHandler(name string, level Level, formatter Formatter) *StdioHandler {
	return &StdioHandler{
		level:      level,
		formatter:  formatter,
		loggerName: name,
		mx:         &sync.Mutex{},
	}
}

func NewDefaultStdioHandler() *StdioHandler {
	return NewStdioHandler("", LevelInfo, NewDefaultLogFormatter())
}

func (h *StdioHandler) String() string {
	return fmt.Sprintf("StdioHandler(%s)", h.loggerName)
}

func (h *StdioHandler) Close() error {
	return nil
}

func (h *StdioHandler) Log(level Level, format string, args ...interface{}) {
	if level >= h.level {
		h.write(
			h.formatter.Format(level, time.Now(), h.loggerName, format, args...),
		)
	}
}

func (h *StdioHandler) Debug(format string, args ...interface{}) {
	h.Log(LevelDebug, format, args...)
}
func (h *StdioHandler) Info(format string, args ...interface{}) {
	h.Log(LevelInfo, format, args...)
}
func (h *StdioHandler) Warn(format string, args ...interface{}) {
	h.Log(LevelWarn, format, args...)
}
func (h *StdioHandler) Error(format string, args ...interface{}) {
	h.Log(LevelError, format, args...)
}

func (h *StdioHandler) write(format string, args ...interface{}) {
	h.mx.Lock()
	defer h.mx.Unlock()
	fmt.Printf(format+"\n", args...)
}

func (h *StdioHandler) SetLevel(level Level) {
	h.level = level
}

func (h *StdioHandler) SetName(name string) {
	h.loggerName = name
}

func (h *StdioHandler) NewFrom(name string) (Handler, error) {
	return &StdioHandler{
		level:      h.level,
		formatter:  h.formatter,
		loggerName: name,
		mx:         &sync.Mutex{},
	}, nil
}

type logFile struct {
	filePath string
	f        *os.File
}

func (lf *logFile) open() error {
	if lf.filePath == "" {
		return fmt.Errorf("cannot open file. no file specified")
	}
	parent := filepath.Dir(lf.filePath)
	if err := os.MkdirAll(parent, 0644); err != nil {
		return fmt.Errorf("failed to make log file parent directory `%s`: %v", parent, err)
	}

	f, err := os.OpenFile(lf.filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file `%s`: %v", lf.filePath, err)
	}
	lf.f = f
	return nil
}

func (lf *logFile) close() error {
	if lf.f != nil {
		return lf.f.Close()
	}
	return nil
}

type FileHandler struct {
	loggerName string
	level      Level
	formatter  Formatter
	file       logFile
	mx         *sync.RWMutex
}

func NewFileHandler(name, filePath string, level Level, formatter Formatter) (*FileHandler, error) {

	fp := filePath
	var err error
	if !filepath.IsAbs(filePath) {
		fp, err = filepath.Abs(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for `%s`: %v", filePath, err)
		}
	}

	file := logFile{filePath: fp}
	if err = file.open(); err != nil {
		return nil, fmt.Errorf("failed to open log file `%s`: %v", filePath, err)
	}

	return &FileHandler{
		level:      level,
		formatter:  formatter,
		loggerName: name,
		file:       file,
		mx:         &sync.RWMutex{},
	}, nil
}

func NewDefaultFileHandler(fileName string) (*FileHandler, error) {
	return NewFileHandler("", fileName, LevelInfo, DefaultFormatter{
		LevelNameMap: map[Level]string{
			LevelDebug: "DEBUG",
			LevelInfo:  "INFO",
			LevelWarn:  "WARN",
			LevelError: "ERROR",
		},
		TimeFormat:     time.RFC3339,
		LevelColourMap: map[Level]string{},
	})
}

func (h *FileHandler) String() string {
	return fmt.Sprintf("FileHandler(%s)", h.loggerName)
}

func (h *FileHandler) Close() error {
	if err := h.file.close(); err != nil {
		return fmt.Errorf("failed to close handle to log file `%s`: %v", h.file.filePath, err)
	}
	return nil
}

func (h *FileHandler) Log(level Level, format string, args ...interface{}) {
	if level >= h.level {
		h.write(
			h.formatter.Format(level, time.Now(), h.loggerName, format, args...),
		)
	}
}

func (h *FileHandler) Debug(format string, args ...interface{}) {
	h.Log(LevelDebug, format, args...)
}
func (h *FileHandler) Info(format string, args ...interface{}) {
	h.Log(LevelInfo, format, args...)
}
func (h *FileHandler) Warn(format string, args ...interface{}) {
	h.Log(LevelWarn, format, args...)
}
func (h *FileHandler) Error(format string, args ...interface{}) {
	h.Log(LevelError, format, args...)
}

func (h *FileHandler) write(format string, args ...interface{}) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.file.f.WriteString(fmt.Sprintf(format+"\n", args...))
	h.file.f.Sync()
}

func (h *FileHandler) SetLevel(level Level) {
	h.level = level
}

func (h *FileHandler) SetName(name string) {
	h.loggerName = name
}

func (h *FileHandler) NewFrom(name string) (Handler, error) {

	file := logFile{filePath: h.file.filePath}
	if err := file.open(); err != nil {
		return nil, fmt.Errorf("failed to open log file `%s`: %v", h.file.filePath, err)
	}

	return &FileHandler{
		level:      h.level,
		formatter:  h.formatter,
		loggerName: name,
		file:       file,
		mx:         h.mx,
	}, nil
}
