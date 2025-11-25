package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// Logger is a structured JSON logger
type Logger struct {
	level       Level
	output      io.Writer
	serviceName string
	environment string
	context     map[string]any
	mu          sync.Mutex
}

// Config holds logger configuration
type Config struct {
	Level       string
	Environment string
	ServiceName string
}

// New creates a new Logger instance
func New(cfg Config) *Logger {
	return &Logger{
		level:       parseLevel(cfg.Level),
		output:      os.Stdout,
		serviceName: cfg.ServiceName,
		environment: cfg.Environment,
		context:     make(map[string]any),
	}
}

// parseLevel converts string to Level
func parseLevel(level string) Level {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// log is the internal logging method
func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Build log entry
	entry := map[string]any{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"level":       level.String(),
		"message":     msg,
		"service":     l.serviceName,
		"environment": l.environment,
	}

	// Add context fields
	for k, v := range l.context {
		entry[k] = v
	}

	// Add additional fields
	for k, v := range fields {
		entry[k] = v
	}

	// Add caller info for errors
	if level >= ErrorLevel {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			entry["caller"] = fmt.Sprintf("%s:%d", file, line)
		}
	}

	// Encode to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	// Write to output
	l.output.Write(data)
	l.output.Write([]byte("\n"))

	// Exit on fatal
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

// DebugFields logs debug with structured fields
func (l *Logger) DebugFields(msg string, fields map[string]any) {
	l.log(DebugLevel, msg, fields)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

// InfoFields logs info with structured fields
func (l *Logger) InfoFields(msg string, fields map[string]any) {
	l.log(InfoLevel, msg, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

// WarnFields logs warning with structured fields
func (l *Logger) WarnFields(msg string, fields map[string]any) {
	l.log(WarnLevel, msg, fields)
}

// Error logs an error message
func (l *Logger) Error(err error) {
	msg := err.Error()
	l.log(ErrorLevel, msg, nil)
}

// ErrorFields logs error with structured fields
func (l *Logger) ErrorFields(msg string, err error, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.log(ErrorLevel, msg, fields)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.log(FatalLevel, msg, nil)
}

// FatalFields logs fatal with structured fields and exits
func (l *Logger) FatalFields(msg string, err error, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.log(FatalLevel, msg, fields)
}

// WithContext returns a new logger with additional context
func (l *Logger) WithContext(fields map[string]any) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newContext := make(map[string]any)
	for k, v := range l.context {
		newContext[k] = v
	}
	for k, v := range fields {
		newContext[k] = v
	}

	return &Logger{
		level:       l.level,
		output:      l.output,
		serviceName: l.serviceName,
		environment: l.environment,
		context:     newContext,
	}
}

// WithTraceID returns a logger with request_id
func (l *Logger) WithTraceID(traceID string) *Logger {
	return l.WithContext(map[string]any{
		"trace_id": traceID,
	})
}

// WithUserID returns a logger with user_id
func (l *Logger) WithUserID(userID string) *Logger {
	return l.WithContext(map[string]any{
		"user_id": userID,
	})
}

// WithError returns a logger with error context
func (l *Logger) WithError(err error) *Logger {
	return l.WithContext(map[string]any{
		"error": err.Error(),
	})
}

// HTTP logs HTTP-related events
func (l *Logger) HTTP(method, path string, statusCode int, duration time.Duration, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	fields["method"] = method
	fields["path"] = path
	fields["status_code"] = statusCode
	fields["duration_ms"] = duration.Milliseconds()

	l.log(InfoLevel, "HTTP request", fields)
}

// responseWriter wraps http.ResponseWriter to capture status code
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
	Error      error
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Body:           &bytes.Buffer{},
	}
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.Body.Write(b)
	return rw.ResponseWriter.Write(b)
}
