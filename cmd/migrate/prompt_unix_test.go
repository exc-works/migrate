//go:build !windows

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
)

func TestTerminalSecretHelper(t *testing.T) {
	mode := os.Getenv("MIGRATE_TEST_TERMINAL_HELPER")
	if mode == "" {
		return
	}
	ctx := context.Background()
	if mode == "cancel" || mode == "dumb" {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
	}
	if mode == "signal" {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(ctx, syscall.SIGTERM)
		defer stop()
	}
	output := os.Stderr
	markerOutput := os.Stdout
	if mode == "stderr" {
		markerOutput = os.Stderr
	}
	secret, err := terminalSecret(ctx, output, "SSH password")
	if mode == "cancel" || mode == "broken" || mode == "dumb" || mode == "ctrl_c" || mode == "ctrl_d" || mode == "signal" || mode == "notty" {
		if err == nil {
			os.Exit(3)
		}
		fmt.Fprintln(os.Stdout, "canceled")
	} else {
		want := os.Getenv("MIGRATE_TEST_TERMINAL_EXPECTED")
		if want == "" {
			want = "example value"
		}
		if err != nil || string(secret) != want {
			os.Exit(4)
		}
		fmt.Fprintln(os.Stdout, "accepted")
	}
	fmt.Fprintln(markerOutput, "restored")
	line, e := bufio.NewReader(os.Stdin).ReadString('\n')
	if e != nil || line != "echo-check\n" {
		os.Exit(6)
	}
	os.Exit(0)
}

func TestTerminalNoEchoAndCancellation(t *testing.T) {
	for _, tc := range []struct{ mode, input, want string }{
		{"input", "example value\n", "example value"},
		{"stderr", "example value\r", "example value"},
		{"spaces", " example value \r", " example value "},
		{"unicode", "wrong\x15密碼🙂x\x7f value\r", "密碼🙂 value"},
		{"ctrl_c", "\x03", ""}, {"ctrl_d", "\x04", ""},
		{"cancel", "", ""}, {"dumb", "", ""},
		{"signal", "", ""}, {"broken", "", ""},
	} {
		mode := tc.mode
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTerminalSecretHelper$")
			cmd.Env = append(os.Environ(), "MIGRATE_TEST_TERMINAL_HELPER="+mode, "MIGRATE_TEST_TERMINAL_EXPECTED="+tc.want)
			if mode == "dumb" {
				cmd.Env = append(cmd.Env, "TERM=dumb")
			}
			var stdout bytes.Buffer
			if mode == "stderr" {
				cmd.Stdout = &stdout
			}
			if mode == "broken" {
				reader, writer, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				reader.Close()
				defer writer.Close()
				cmd.Stderr = writer
			}
			terminal, err := pty.Start(cmd)
			if err != nil {
				t.Fatal(err)
			}
			defer terminal.Close()
			var output bytes.Buffer
			var one [1]byte
			marker := "SSH password: "
			if mode == "broken" {
				marker = "restored"
			}
			for !strings.Contains(output.String(), marker) {
				n, err := terminal.Read(one[:])
				if err != nil {
					t.Fatal(err)
				}
				output.Write(one[:n])
			}
			if tc.input != "" {
				if _, err := io.WriteString(terminal, tc.input); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "signal" {
				if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
					t.Fatal(err)
				}
			}
			for !strings.Contains(output.String(), "restored") {
				n, err := terminal.Read(one[:])
				if err != nil {
					t.Fatal(err)
				}
				output.Write(one[:n])
			}
			if _, err := io.WriteString(terminal, "echo-check\n"); err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(&output, terminal)
			if !strings.Contains(output.String(), "echo-check") {
				t.Fatal("terminal echo was not restored")
			}
			if err := cmd.Wait(); err != nil {
				t.Fatalf("helper failed: %v output=%q", err, output.String())
			}
			if (tc.want != "" && strings.Contains(output.String(), tc.want)) || strings.Contains(output.String(), "wrong") {
				t.Fatal("password echoed")
			}
			if mode == "stderr" {
				if stdout.String() != "accepted\n" {
					t.Fatalf("prompt polluted stdout: %q", stdout.String())
				}
			} else {
				marker := "canceled"
				if tc.want != "" {
					marker = "accepted"
				}
				if !strings.Contains(output.String(), marker) {
					t.Fatalf("unexpected output: %q", output.String())
				}
			}
		})
	}
}

func TestTerminalSecretWithoutControllingTerminal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTerminalSecretHelper$")
	cmd.Env = append(os.Environ(), "MIGRATE_TEST_TERMINAL_HELPER=notty")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdin = strings.NewReader("echo-check\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("no-terminal helper: %v output=%q", err, output)
	}
	if strings.Contains(string(output), "SSH password") || !strings.Contains(string(output), "canceled") {
		t.Fatalf("no-terminal input was prompted: %q", output)
	}
}
