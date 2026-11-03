//go:build !windows

package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func openSecretTerminal() (*os.File, error) {
	path := filepath.Join(string(os.PathSeparator), "dev", "tty")
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), "terminal"), nil
}

func readSecretTerminal(ctx context.Context, file *os.File) ([]byte, error) {
	fd := int(file.Fd())
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		poll := []unix.PollFd{{
			Fd:     int32(fd),
			Events: unix.POLLIN,
		}}
		n, err := unix.Poll(poll, 100)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return nil, errors.New("terminal input failed")
		}
		if n == 0 {
			continue
		}
		var buf [128]byte
		count, err := unix.Read(fd, buf[:])
		if err == unix.EAGAIN || err == unix.EINTR {
			continue
		}
		if err != nil || count == 0 {
			return nil, errors.New("terminal input ended")
		}
		return buf[:count], nil
	}
}

// A broken stderr pipe must return EPIPE so terminal restoration can run.
func protectPromptOutput() func() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGPIPE)
	return func() { signal.Stop(ch) }
}
