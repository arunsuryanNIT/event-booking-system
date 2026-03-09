// Package logger provides a structured JSON logger that writes to stdout or a file.
// Log entries are emitted as single-line JSON objects for easy ingestion by log
// aggregators (ELK, CloudWatch, Loki, etc.).
//
// Usage:
//
//	l := logger.New(logger.Options{Output: "stdout", Level: "info"})
//	l.Info("server started", "port", "8080", "env", "production")
//
// Produces:
//
//	{"timestamp":"2025-08-15T10:00:00Z","level":"INFO","message":"server started","port":"8080","env":"production"}
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents log severity.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

// ParseLevel converts a string level name to a Level.
// Defaults to LevelInfo for unrecognised values.
func ParseLevel(s string) Level {
	switch s {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO", "":
		return LevelInfo
	case "warn", "WARN":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger writes structured JSON log entries to a configured writer.
// It is safe for concurrent use.
type Logger struct {
	mu    sync.Mutex
	w     io.Writer
	level Level
}

// Options configures a new Logger.
type Options struct {
	// Output is the log destination: "stdout", "stderr", or a file path.
	// Defaults to "stdout" if empty.
	Output string

	// Level is the minimum severity to emit: "debug", "info", "warn", "error".
	// Defaults to "info" if empty.
	Level string
}

// New creates a Logger from the given Options.
// If Output specifies a file path, the file is opened in append mode (created if needed).
func New(opts Options) *Logger {
	var w io.Writer

	switch opts.Output {
	case "", "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	default:
		f, err := os.OpenFile(opts.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "logger: failed to open %s, falling back to stdout: %v\n", opts.Output, err)
			w = os.Stdout
		} else {
			w = f
		}
	}

	return &Logger{
		w:     w,
		level: ParseLevel(opts.Level),
	}
}

// entry is the JSON structure written for each log line.
type entry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// log writes a single JSON log line if level >= the configured minimum.
// Extra key-value pairs are added as top-level JSON fields.
// Keys must be strings; values are serialised as-is by encoding/json.
func (l *Logger) log(level Level, msg string, kvs ...interface{}) {
	if level < l.level {
		return
	}

	m := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     levelNames[level],
		"message":   msg,
	}

	for i := 0; i+1 < len(kvs); i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", kvs[i])
		}
		m[key] = kvs[i+1]
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, _ := json.Marshal(m)
	data = append(data, '\n')
	l.w.Write(data)
}

// Debug logs at DEBUG level.
func (l *Logger) Debug(msg string, kvs ...interface{}) {
	l.log(LevelDebug, msg, kvs...)
}

// Info logs at INFO level.
func (l *Logger) Info(msg string, kvs ...interface{}) {
	l.log(LevelInfo, msg, kvs...)
}

// Warn logs at WARN level.
func (l *Logger) Warn(msg string, kvs ...interface{}) {
	l.log(LevelWarn, msg, kvs...)
}

// Error logs at ERROR level.
func (l *Logger) Error(msg string, kvs ...interface{}) {
	l.log(LevelError, msg, kvs...)
}
