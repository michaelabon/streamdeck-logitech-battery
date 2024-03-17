//go:build darwin

package main

import (
	"log"
	"os"
	"path/filepath"
)

func getDbFilepath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("Unable to obtain user’s home directory", err)

		return "", err
	}

	dbFilePath := filepath.Join(home, "Library", "Application Support", "LGHUB", "settings.db")

	return dbFilePath, nil
}
