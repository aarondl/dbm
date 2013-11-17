package main

import (
	"github.com/aarondl/paths"
	"io"
	"os"
	"path/filepath"
)

const basicConfig = `[development]
name = "development"
kind = "sqlite3"

[testing]
name = "testing"
kind = "sqlite3"

[production]
host = "myserver.com" # Could also be "/var/run/mysqld/mysqld.sock"
name = "production"
kind = "mysql"
user = "username"
pass = "password"
`

func initialize(args []string) {
	configDir := filepath.Join(workingDir, _DATA_DIR)
	configFile := filepath.Join(configDir, _CONFIG)
	if _, err := paths.EnsureDirectory(configDir); err != nil {
		exitLn("Error verifying dir:", err)
	}

	if ok, err := paths.FileExists(configFile); err != nil {
		exitLn("Error verifying config:", err)
	} else if ok {
		return
	}

	var f *os.File
	var err error
	if f, err = os.Create(configFile); err != nil {
		exitLn("Could not create config:", err)
	}
	defer f.Close()

	if _, err := io.WriteString(f, basicConfig); err != nil {
		exitLn("Could not create config:", err)
	}
}
