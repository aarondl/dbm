package main

import (
	"github.com/aarondl/dbm/config"
	"path/filepath"
)

func initialize(args []string) {
	configDir := filepath.Join(workingDir, config.DATA_DIR)
	if err := config.Touch(configDir); err != nil {
		exitLn("Error creating config:", err)
	}
}
