package config

import (
	"testing"
	"time"
)

func TestScriptConfiguration(t *testing.T) {
	cfg := FileConfig{ScriptTimeout: "30s", ScriptMaxRows: 42, ScriptMaxBytes: 4096}
	options, err := cfg.ScriptOptions()
	if err != nil || options.Timeout != 30*time.Second || options.MaxRows != 42 || options.MaxBytes != 4096 {
		t.Fatalf("%+v %v", options, err)
	}
	for _, cfg := range []FileConfig{{ScriptTimeout: "invalid"}, {ScriptTimeout: "0s"}, {ScriptTimeout: "-1s"}, {ScriptMaxRows: -1}, {ScriptMaxBytes: -1}} {
		if _, err = cfg.ScriptOptions(); err == nil {
			t.Fatalf("accepted %+v", cfg)
		}
	}
}
