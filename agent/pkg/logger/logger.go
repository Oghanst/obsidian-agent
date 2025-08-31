package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

type Logger struct {
	logger *log.Logger
}

// New creates a logger writing to both file and stdout
func New(logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// MultiWriter -> file + stdout
	writer := io.MultiWriter(file, os.Stdout)

	return &Logger{
		logger: log.New(writer, "", 0), // no default prefix
	}, nil
}

// internal log function
func (l *Logger) log(level string, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// get caller info (skip=3 保证拿到调用者，而不是 log 包本身)
	_, file, line, ok := runtime.Caller(2)
	loc := "unknown"
	if ok {
		loc = fmt.Sprintf("%s:%d", file, line)
	}

	// format final log line
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] [%s] %s", timestamp, level, loc, message)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log("DEBUG", format, args...)
}
