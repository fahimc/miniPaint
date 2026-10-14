//go:build !windows

package main

import (
	"os/exec"
	"runtime"
)

func openBrowser(address string) {
	program := "xdg-open"
	if runtime.GOOS == "darwin" {
		program = "open"
	}
	_ = exec.Command(program, address).Start()
}
