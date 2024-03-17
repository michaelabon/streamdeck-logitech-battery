//go:build windows

package main

import (
	"log"
	"os"
	"path/filepath"
)

func getDbFilepath() (string, error) {
	cacheDir, err := os.UserCacheDir() // %LOCALAPPDATA%
	if err != nil {
		log.Println("Unable to obtain user’s cache directory", err)

		return "", err
	}

	dbFilePath := filepath.Join(cacheDir, "LGHUB", "settings.db")

	return dbFilePath, nil
}
