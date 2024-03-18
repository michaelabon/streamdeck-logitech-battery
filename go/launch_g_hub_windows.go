//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func launchGHub() error {
	cmd := open("GHub.exe")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error while opening G Hub: %w", err)
	}

	return nil
}

func open(input string) *exec.Cmd {
	runDll32 := filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")

	return exec.Command(runDll32, "url.dll,FileProtocolHandler", input)
}
