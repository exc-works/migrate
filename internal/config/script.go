package config

import (
	"errors"
	"time"

	"github.com/exc-works/migrate/internal/script"
)

// ScriptOptions validates CLI limits before opening a database.
func (cfg FileConfig) ScriptOptions() (script.Options, error) {
	options := script.Options{MaxRows: cfg.ScriptMaxRows, MaxBytes: cfg.ScriptMaxBytes}
	if cfg.ScriptTimeout != "" {
		duration, err := time.ParseDuration(cfg.ScriptTimeout)
		if err != nil || duration <= 0 {
			return options, errors.New("script_timeout must be a positive duration")
		}
		options.Timeout = duration
	}
	return options.Normalize()
}
