package main

import (
	"context"
	"errors"
	"os"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var consoleEvents = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetNumberOfConsoleInputEvents")
var consoleRead = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReadConsoleInputW")

func openSecretTerminal() (*os.File, error) { return os.OpenFile("CONIN$", os.O_RDWR, 0) }

func readSecretTerminal(ctx context.Context, file *os.File) ([]byte, error) {
	var high uint16
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var count uint32
		ok, _, _ := consoleEvents.Call(file.Fd(), uintptr(unsafe.Pointer(&count)))
		if ok == 0 {
			return nil, errors.New("cannot inspect terminal input")
		}
		if count == 0 {
			timer := time.NewTimer(50 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
			continue
		}
		// INPUT_RECORD is a type tag plus a 16-byte union; KEY_EVENT_RECORD's
		// key-down and Unicode character fields are at offsets 4 and 14.
		var record struct {
			Kind    uint16
			Padding uint16
			Down    int32
			Repeat  uint16
			Key     uint16
			Scan    uint16
			Char    uint16
			Control uint32
		}
		var read uint32
		ok, _, _ = consoleRead.Call(file.Fd(), uintptr(unsafe.Pointer(&record)), 1, uintptr(unsafe.Pointer(&read)))
		if ok == 0 {
			return nil, errors.New("terminal input failed")
		}
		if read == 0 || record.Kind != 1 || record.Down == 0 || record.Char == 0 {
			continue
		}
		if record.Char >= 0xD800 && record.Char <= 0xDBFF {
			high = record.Char
			continue
		}
		ch := rune(record.Char)
		if high != 0 {
			ch = utf16.DecodeRune(rune(high), ch)
		}
		return repeatConsoleRune(ch, record.Repeat), nil
	}
}

func protectPromptOutput() func() { return func() {} }
