package logger

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type LogBuffer struct {
	buffer   []LogEntry
	maxSize  int
	mu       sync.RWMutex
	channels map[chan LogEntry]bool
	chanMu   sync.RWMutex
}

var globalLogBuffer *LogBuffer

func InitLogBuffer(maxSize int) *LogBuffer {
	globalLogBuffer = &LogBuffer{
		buffer:   make([]LogEntry, 0, maxSize),
		maxSize:  maxSize,
		channels: make(map[chan LogEntry]bool),
	}
	return globalLogBuffer
}

func GetLogBuffer() *LogBuffer {
	return globalLogBuffer
}

func (lb *LogBuffer) AddLog(level, message string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}

	lb.mu.Lock()
	lb.buffer = append(lb.buffer, entry)
	if len(lb.buffer) > lb.maxSize {
		lb.buffer = lb.buffer[len(lb.buffer)-lb.maxSize:]
	}
	lb.mu.Unlock()

	lb.chanMu.RLock()
	for ch := range lb.channels {
		select {
		case ch <- entry:
		default:
		}
	}
	lb.chanMu.RUnlock()
}

func (lb *LogBuffer) GetLogs(count int) []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if count <= 0 || count > len(lb.buffer) {
		return append([]LogEntry{}, lb.buffer...)
	}

	start := len(lb.buffer) - count
	return append([]LogEntry{}, lb.buffer[start:]...)
}

func (lb *LogBuffer) Subscribe() chan LogEntry {
	ch := make(chan LogEntry, 100)
	lb.chanMu.Lock()
	lb.channels[ch] = true
	lb.chanMu.Unlock()
	return ch
}

func (lb *LogBuffer) Unsubscribe(ch chan LogEntry) {
	lb.chanMu.Lock()
	delete(lb.channels, ch)
	close(ch)
	lb.chanMu.Unlock()
}

func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	lb.buffer = make([]LogEntry, 0, lb.maxSize)
	lb.mu.Unlock()
}

type LogWriter struct {
	level string
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	message := string(bytes.TrimSpace(p))
	if message == "" {
		return len(p), nil
	}

	if globalLogBuffer != nil {
		globalLogBuffer.AddLog(lw.level, message)
	}

	return len(p), nil
}

func NewLogWriter(level string) io.Writer {
	return &LogWriter{level: level}
}

func Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if globalLogBuffer != nil {
		globalLogBuffer.AddLog("INFO", message)
	}
	fmt.Printf("[INFO] %s\n", message)
}

func Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if globalLogBuffer != nil {
		globalLogBuffer.AddLog("ERROR", message)
	}
	fmt.Printf("[ERROR] %s\n", message)
}

func Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if globalLogBuffer != nil {
		globalLogBuffer.AddLog("DEBUG", message)
	}
}

func Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	if globalLogBuffer != nil {
		globalLogBuffer.AddLog("WARN", message)
	}
	fmt.Printf("[WARN] %s\n", message)
}
