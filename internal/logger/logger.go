package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger provides a simple logging interface
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugMode   bool
	logFile     *os.File
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

// InitFileLogging initializes logging to a file
func (l *Logger) InitFileLogging(logPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file (append mode)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	l.logFile = f

	// Redirect all loggers to the file
	l.debugLogger.SetOutput(f)
	l.infoLogger.SetOutput(f)
	l.warnLogger.SetOutput(f)
	l.errorLogger.SetOutput(f)

	// Also redirect standard logger
	log.SetOutput(f)

	return nil
}

// Close closes the log file if it's open
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
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
