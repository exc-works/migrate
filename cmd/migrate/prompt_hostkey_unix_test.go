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
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/exc-works/migrate/internal/tunnel"
)

func TestHostKeyTerminalHelper(t *testing.T) {
	mode := os.Getenv("MIGRATE_TEST_HOSTKEY_HELPER")
	if mode == "" {
		return
	}
	ctx := context.Background()
	if mode == "cancel" {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
	}
	approved, err := terminalHostKey(ctx, os.Stderr, tunnel.HostKeyInfo{Address: "ssh.internal.example:2222", Algorithm: "ssh-ed25519", Fingerprint: "SHA256:example"})
	if mode == "cancel" || mode == "broken" {
		if err == nil || approved {
			os.Exit(3)
		}
	} else if err != nil || approved != (mode == "yes" || mode == "stderr") {
		os.Exit(4)
	}
	marker := os.Stdout
	if mode == "stderr" {
		marker = os.Stderr
	}
	fmt.Fprintln(os.Stdout, `{"result":"finished"}`)
	fmt.Fprintln(marker, "restored")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || line != "echo-check\n" {
		os.Exit(5)
	}
	os.Exit(0)
}

func TestHostKeyTerminalConfirmation(t *testing.T) {
	for _, tc := range []struct{ mode, input string }{{"yes", "y"}, {"no", "n"}, {"default", "\r"}, {"cancel", ""}, {"stderr", "y"}, {"broken", ""}} {
		t.Run(tc.mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHostKeyTerminalHelper$")
			cmd.Env = append(os.Environ(), "MIGRATE_TEST_HOSTKEY_HELPER="+tc.mode)
			var stdout bytes.Buffer
			if tc.mode == "stderr" {
				cmd.Stdout = &stdout
			}
			if tc.mode == "broken" {
				reader, writer, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				reader.Close()
				defer writer.Close()
				cmd.Stderr = writer
			}
			terminal, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 24, Cols: 100})
			if err != nil {
				t.Fatal(err)
			}
			defer terminal.Close()
			var output bytes.Buffer
			marker := "SHA256:example"
			if tc.mode == "broken" {
				marker = "restored"
			}
			var one [1]byte
			for !strings.Contains(output.String(), marker) {
				n, err := terminal.Read(one[:])
				if err != nil {
					t.Fatalf("prompt: %v output=%q", err, output.String())
				}
				output.Write(one[:n])
			}
			if _, err := io.WriteString(terminal, tc.input); err != nil {
				t.Fatal(err)
			}
			for !strings.Contains(output.String(), "restored") {
				n, err := terminal.Read(one[:])
				if err != nil {
					t.Fatalf("confirmation: %v output=%q", err, output.String())
				}
				output.Write(one[:n])
			}
			if _, err := io.WriteString(terminal, "echo-check\n"); err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(&output, terminal)
			if err := cmd.Wait(); err != nil {
				t.Fatalf("helper: %v output=%q", err, output.String())
			}
			if !strings.Contains(output.String(), "echo-check") {
				t.Fatal("terminal echo not restored")
			}
			if tc.mode != "broken" && (!strings.Contains(output.String(), "ssh.internal.example:2222") || !strings.Contains(output.String(), "ssh-ed25519")) {
				t.Fatal("host identity not displayed")
			}
			if tc.mode == "stderr" && stdout.String() != "{\"result\":\"finished\"}\n" {
				t.Fatalf("prompt polluted stdout: %q", stdout.String())
			}
		})
	}
}
