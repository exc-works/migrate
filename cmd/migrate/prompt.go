package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"golang.org/x/term"
)

// terminalSecret owns terminal state until it returns, including cancellation.
// Raw reads are polled so Ctrl+C/context cancellation cannot leave a blocked reader.
func terminalSecret(ctx context.Context, out io.Writer, label string) (value []byte, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := openSecretTerminal()
	if err != nil {
		return nil, errors.New("no terminal available; set the secret environment variable or use non-interactive mode")
	}
	defer file.Close()
	stopSignals := protectPromptOutput()
	defer stopSignals()
	state, err := term.MakeRaw(int(file.Fd()))
	if err != nil {
		return nil, errors.New("cannot disable terminal echo")
	}
	defer func() {
		if restoreErr := term.Restore(int(file.Fd()), state); restoreErr != nil {
			err = errors.Join(err, errors.New("cannot restore terminal state"))
		}
		_, _ = fmt.Fprintln(out)
	}()
	if _, err := fmt.Fprintf(out, "%s: ", label); err != nil {
		return nil, errors.New("cannot write secret prompt")
	}
	for {
		data, err := readSecretTerminal(ctx, file)
		if err != nil {
			return nil, err
		}
		for _, ch := range data {
			switch ch {
			case 3, 4:
				return nil, errors.New("secret input canceled")
			case '\r', '\n':
				return value, nil
			case 8, 127:
				if len(value) > 0 {
					_, size := utf8.DecodeLastRune(value)
					value = value[:len(value)-size]
				}
			case 21:
				value = value[:0]
			default:
				if ch >= 32 {
					value = append(value, ch)
				}
				if len(value) > 65536 {
					return nil, errors.New("secret input is too long")
				}
			}
		}
	}
}

func repeatConsoleRune(ch rune, count uint16) []byte {
	var data []byte
	for i := uint16(0); i < count; i++ {
		data = utf8.AppendRune(data, ch)
	}
	return data
}
