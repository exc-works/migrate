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
)

func TestTerminalSecretHelper(t *testing.T) {
	mode := os.Getenv("MIGRATE_TEST_TERMINAL_HELPER")
	if mode == "" {
		return
	}
	ctx := context.Background()
	if mode == "cancel" {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
	}
	output := os.Stdout
	if mode == "broken" {
		output = os.Stderr
	}
	secret, err := terminalSecret(ctx, output, "SSH password")
	if mode == "cancel" || mode == "broken" {
		if err == nil {
			os.Exit(3)
		}
		fmt.Fprintln(os.Stdout, "canceled")
	} else {
		if err != nil || string(secret) != "example value" {
			os.Exit(4)
		}
		fmt.Fprintln(os.Stdout, "accepted")
	}
	fmt.Fprintln(os.Stdout, "restored")
	line, e := bufio.NewReader(os.Stdin).ReadString('\n')
	if e != nil || line != "echo-check\n" {
		os.Exit(6)
	}
	os.Exit(0)
}

func TestTerminalNoEchoAndCancellation(t *testing.T) {
	for _, mode := range []string{"input", "cancel", "broken"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTerminalSecretHelper$")
			cmd.Env = append(os.Environ(), "MIGRATE_TEST_TERMINAL_HELPER="+mode)
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
			if mode == "input" {
				if _, err := io.WriteString(terminal, "example value\n"); err != nil {
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
			if strings.Contains(output.String(), "example value") {
				t.Fatal("password echoed")
			}
			if !strings.Contains(output.String(), map[string]string{"input": "accepted", "cancel": "canceled", "broken": "canceled"}[mode]) {
				t.Fatalf("unexpected output: %q", output.String())
			}
		})
	}
}
