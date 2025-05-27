package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides a simple logging interface
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugMode   bool
}

// New creates a new Logger instance
func New() *Logger {
	return &Logger{
		debugLogger: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime),
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		warnLogger:  log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
		debugMode:   false,
	}
}

// NewWithDebug creates a new Logger instance with debug mode enabled
func NewWithDebug() *Logger {
	logger := New()
	logger.debugMode = true
	return logger
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

// Debug logs a debug message (only when debug mode is enabled)
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.debugMode {
		l.debugLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// SetDebugMode enables or disables debug logging
func (l *Logger) SetDebugMode(enabled bool) {
	l.debugMode = enabled
}

// IsDebugMode returns whether debug mode is enabled
func (l *Logger) IsDebugMode() bool {
	return l.debugMode
}

// LogTiming is a helper to log the time taken for a function to execute
func (l *Logger) LogTiming(name string) func() {
	start := time.Now()
	return func() {
		l.Info("%s took %v", name, time.Since(start))
	}
}
