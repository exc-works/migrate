//go:build !windows

package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func openSecretTerminal() (*os.File, error) {
	path := filepath.Join(string(os.PathSeparator), "dev", "tty")
	return os.OpenFile(path, os.O_RDWR, 0)
}

// A broken stderr pipe must return EPIPE so Huh can restore the terminal.
func protectPromptOutput() func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGPIPE)
	return func() { signal.Stop(ch) }
}
