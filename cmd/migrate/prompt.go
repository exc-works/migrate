package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

const maxSecretBytes = 64 * 1024

// terminalSecret delegates editing, hidden input and terminal restoration to Huh.
func terminalSecret(ctx context.Context, out io.Writer, label string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// A broken stderr pipe must return EPIPE so Bubble Tea can restore the
	// terminal before returning. Windows does not generate SIGPIPE.
	if runtime.GOOS != "windows" {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGPIPE)
		defer signal.Stop(signals)
	}
	promptCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	output := &secretPromptOutput{
		Writer: out,
		cancel: cancel,
	}
	var value string
	keys := huh.NewDefaultKeyMap()
	keys.Quit.SetKeys("ctrl+c", "ctrl+d")
	keys.Input.Submit.SetKeys("enter", "ctrl+j")
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title(label).Inline(true).Prompt(": ").
			EchoMode(huh.EchoModeNone).Value(&value),
	)).
		WithShowHelp(false).
		WithKeyMap(keys)
	form.SubmitCmd = tea.Quit
	form.CancelCmd = tea.Quit
	program := tea.NewProgram(form, tea.WithInputTTY(), tea.WithOutput(output), tea.WithoutSignalHandler())
	// Run Huh as a Bubble Tea model so cancellation uses graceful Quit. Its
	// RunWithContext kills the program and can close the reader before it exits.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-promptCtx.Done():
			program.Quit()
		case <-done:
		}
	}()
	_, err := program.Run()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil || promptCtx.Err() != nil {
		return nil, errors.New("secret input failed; set the secret environment variable or use non-interactive mode")
	}
	if form.State != huh.StateCompleted {
		return nil, errors.New("secret input canceled")
	}
	if len(value) > maxSecretBytes {
		return nil, errors.New("secret input is too long")
	}
	return []byte(value), nil
}

// Bubble Tea ignores renderer write errors; cancel the form if stderr closes.
// Preserve Fd so its terminal setup still works on Windows consoles.
type secretPromptOutput struct {
	io.Writer
	cancel context.CancelFunc
}

func (w *secretPromptOutput) Write(data []byte) (int, error) {
	n, err := w.Writer.Write(data)
	if err != nil {
		w.cancel()
	}
	return n, err
}

func (w *secretPromptOutput) Fd() uintptr {
	if file, ok := w.Writer.(interface{ Fd() uintptr }); ok {
		return file.Fd()
	}
	return ^uintptr(0)
}
