package migrate

import (
	"io"

	"github.com/exc-works/migrate/internal/logger"
)

// Logger is the minimal leveled logging interface Service writes to.
type Logger = logger.Logger

// NoopLogger discards all log output. Use as the default when embedding.
type NoopLogger = logger.NoopLogger

// StdLogger writes leveled output through the standard library log package.
type StdLogger = logger.StdLogger

// LogLevel is the threshold used by StdLogger and returned by ParseLogLevel.
type LogLevel = logger.Level

// Log thresholds for StdLogger; messages below the configured level are dropped.
const (
	LogLevelDebug = logger.LevelDebug
	LogLevelInfo  = logger.LevelInfo
	LogLevelWarn  = logger.LevelWarn
	LogLevelError = logger.LevelError
)

// NewStdLogger constructs a StdLogger writing to output at the given threshold string.
func NewStdLogger(level string, output io.Writer) *StdLogger {
	return logger.NewStd(level, output)
}

// ParseLogLevel converts a case-insensitive level string ("debug"/"info"/"warn"/"error") into a LogLevel.
func ParseLogLevel(level string) LogLevel {
	return logger.ParseLevel(level)
}
