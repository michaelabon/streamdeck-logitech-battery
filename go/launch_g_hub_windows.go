//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func launchGHub() error {
	path := filepath.Join("C:\\", "Program Files", "LGHUB", "lghub.exe")
	cmd := exec.Command(path)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error while opening G Hub: %w", err)
	}

	return nil
}
