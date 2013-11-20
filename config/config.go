package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/aarondl/paths"
	"io"
	"os"
	"path/filepath"
)

// DB is a database configuration.
type DB struct {
	Name string
	Kind string
	Host string
	User string
	Pass string
}

const (
	// CONFIG is the name of the configuration file.
	CONFIG = "config.toml"
	// DATA_DIR is the directory from ROOT where the database files live.
	DATA_DIR = "db"

	errVCSRoot = "dbmconfig: Could not determine VCS root to search for config."
	errConfig  = "dbmconfig: Could not find a configuration file to load."
	errEnv     = "dbmconfig: Environment missing - "

	errCreatingConfig = "dbmconfig: Error creating configuration - "
)

// Configuration is the type of the current config.
type Configuration map[string]*DB

var (
	// AllConfigs is all the environments.
	AllConfigs Configuration
	// Current is the currently requested environment.
	Current *DB
)

// LoadFile loads a configuration at a particular path.
func LoadFile(path, env string) error {
	if _, err := toml.DecodeFile(path, &AllConfigs); err != nil {
		return err
	}

	if c, ok := AllConfigs[env]; !ok {
		return errors.New(errEnv + env)
	} else {
		Current = c
	}

	return nil
}

// Load loads a configuration from CWD/db/config.toml first, if it cannot
// find that, it attempts to find the VCS root and do VCSROOT/db/config.toml.
func Load(env string) error {
	var wd string
	var yes bool
	var err error
	if wd, err = os.Getwd(); err != nil {
		return err
	}

	cwdConfig := filepath.Join(wd, DATA_DIR, CONFIG)
	if yes, err = paths.FileExists(cwdConfig); err != nil {
		return err
	} else if yes {
		return LoadFile(cwdConfig, env)
	}

	var vcsRoot string
	if _, vcsRoot, err = paths.FindVCSRoot(wd); err != nil {
		return err
	} else if len(vcsRoot) == 0 {
		return errors.New(errVCSRoot)
	}

	vcsConfig := filepath.Join(vcsRoot, DATA_DIR, CONFIG)
	if yes, err = paths.FileExists(vcsConfig); err != nil {
		return err
	} else if yes {
		return LoadFile(vcsConfig, env)
	}

	return errors.New(errConfig)
}

// Touch creates a basic configuration file. Dir should be a path to where the
// config should be written.
func Touch(dir string) error {
	if _, err := paths.EnsureDirectory(dir); err != nil {
		return errors.New(errCreatingConfig + err.Error())
	}

	configFile := filepath.Join(dir, CONFIG)
	if ok, err := paths.FileExists(configFile); err != nil {
		return errors.New(errCreatingConfig + err.Error())
	} else if ok {
		return nil
	}

	var f *os.File
	var err error
	if f, err = os.Create(configFile); err != nil {
		return errors.New(errCreatingConfig + err.Error())
	}
	defer f.Close()

	if _, err := io.WriteString(f, basicConfig); err != nil {
		return errors.New(errCreatingConfig + err.Error())
	}

	return nil
}

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
