//go:build darwin

package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func launchGHub() error {
	cmd := exec.Command("open", "-b", "com.logi.ghub")
	err := cmd.Run()

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err != nil {
		return fmt.Errorf("error while opening G Hub: %s: %w", stderr.String(), err)
	}

	return nil
}
