// Package script hosts the optional, synchronous migration runtime.
package script

import (
	"errors"
	"time"
)

// Options bounds execution and materialized query results. Zero values use defaults.
type Options struct {
	Timeout  time.Duration
	MaxRows  int
	MaxBytes int64
}

// Normalize validates limits before a migration plan is executed.
func (o Options) Normalize() (Options, error) {
	if o.Timeout < 0 || o.MaxRows < 0 || o.MaxBytes < 0 {
		return o, errors.New("script limits must be positive")
	}
	if o.Timeout == 0 {
		o.Timeout = 5 * time.Minute
	}
	if o.MaxRows == 0 {
		o.MaxRows = 1000
	}
	if o.MaxBytes == 0 {
		o.MaxBytes = 16 << 20
	}
	return o, nil
}

// Build identifiers reported by capabilities.
const (
	BuildSQL  = "sql"
	BuildFull = "full"
)

var ErrRequiresFull = errors.New("JS/TS migrations require a full build (build tag: full)")
