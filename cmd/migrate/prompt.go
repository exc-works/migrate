package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/exc-works/migrate/internal/tunnel"
)

const maxSecretBytes = 64 * 1024

// terminalSecret delegates editing, hidden input and terminal restoration to Huh.
func terminalSecret(ctx context.Context, out io.Writer, label string) ([]byte, error) {
	var value string
	input := huh.NewInput().Title(label).Inline(true).Prompt(": ").
		EchoMode(huh.EchoModeNone).Value(&value)
	if err := runTerminalForm(ctx, out, input); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("secret input canceled or unavailable; set the secret environment variable or use non-interactive mode")
	}
	if len(value) > maxSecretBytes {
		return nil, errors.New("secret input is too long")
	}
	return []byte(value), nil
}

func terminalHostKey(ctx context.Context, out io.Writer, info tunnel.HostKeyInfo) (bool, error) {
	approved := false
	confirmation := huh.NewConfirm().Title("Trust this SSH server?").
		Description(fmt.Sprintf("Address: %s\nKey type: %s\nFingerprint: %s\nVerify this fingerprint before trusting the server.", info.Address, info.Algorithm, info.Fingerprint)).
		Affirmative("Trust and save").Negative("Cancel").Value(&approved)
	if err := runTerminalForm(ctx, out, confirmation); err != nil {
		return false, err
	}
	return approved, nil
}

// runTerminalForm keeps Huh's default theme and delegates all terminal handling.
func runTerminalForm(ctx context.Context, out io.Writer, field huh.Field) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// A broken stderr pipe must return EPIPE so Bubble Tea can restore the terminal.
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
	keys := huh.NewDefaultKeyMap()
	keys.Quit.SetKeys("ctrl+c", "ctrl+d")
	keys.Input.Submit.SetKeys("enter", "ctrl+j")
	keys.Confirm.Submit.SetKeys("enter", "ctrl+j")
	form := huh.NewForm(huh.NewGroup(field)).WithShowHelp(false).WithKeyMap(keys)
	form.SubmitCmd = tea.Quit
	form.CancelCmd = tea.Quit
	program := tea.NewProgram(form, tea.WithInputTTY(), tea.WithOutput(output), tea.WithoutSignalHandler())
	// Use graceful Quit so the input reader exits before terminal cleanup.
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
		return ctx.Err()
	}
	if err != nil || promptCtx.Err() != nil {
		return errors.New("terminal input failed")
	}
	if form.State != huh.StateCompleted {
		return errors.New("terminal input canceled")
	}
	return nil
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
