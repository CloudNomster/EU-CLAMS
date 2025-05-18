package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides a simple logging interface
type Logger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

// New creates a new Logger instance
func New() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		warnLogger:  log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Output(2, fmt.Sprintf(format, v...))
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.warnLogger.Output(2, fmt.Sprintf(format, v...))
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Output(2, fmt.Sprintf(format, v...))
}

// LogTiming is a helper to log the time taken for a function to execute
func (l *Logger) LogTiming(name string) func() {
	start := time.Now()
	return func() {
		l.Info("%s took %v", name, time.Since(start))
	}
}
