package main

import "os"

func openSecretTerminal() (*os.File, error) { return os.OpenFile("CONIN$", os.O_RDWR, 0) }

func protectPromptOutput() func() { return func() {} }
