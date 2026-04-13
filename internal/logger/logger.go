package logger

import (
	"io"
	"log"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type StdLogger struct {
	level Level
	impl  *log.Logger
}

func NewStd(level string, output io.Writer) *StdLogger {
	if output == nil {
		output = os.Stdout
	}
	return &StdLogger{
		level: ParseLevel(level),
		impl:  log.New(output, "", log.LstdFlags),
	}
}

func ParseLevel(level string) Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *StdLogger) logf(level Level, prefix, format string, args ...any) {
	if l == nil || l.impl == nil {
		return
	}
	if level < l.level {
		return
	}
	l.impl.Printf(prefix+format, args...)
}

func (l *StdLogger) Debugf(format string, args ...any) {
	l.logf(LevelDebug, "DEBUG ", format, args...)
}

func (l *StdLogger) Infof(format string, args ...any) {
	l.logf(LevelInfo, "INFO ", format, args...)
}

func (l *StdLogger) Warnf(format string, args ...any) {
	l.logf(LevelWarn, "WARN ", format, args...)
}

func (l *StdLogger) Errorf(format string, args ...any) {
	l.logf(LevelError, "ERROR ", format, args...)
}

type NoopLogger struct{}

func (NoopLogger) Debugf(string, ...any) {}
func (NoopLogger) Infof(string, ...any)  {}
func (NoopLogger) Warnf(string, ...any)  {}
func (NoopLogger) Errorf(string, ...any) {}
