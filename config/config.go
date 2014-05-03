/*
Package config gives access to the dbm configuration file and the ability to
create connection strings in order to connect with the configured databases.
*/
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aarondl/paths"
)

// DB is a database configuration.
type DB struct {
	Name          string
	Kind          string
	Host          string
	User          string
	Pass          string
	SSL           bool
	SSLSkipVerify bool
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
	// Current is the config of the requested environment.
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

// DSN creates a connection string from the database values given.
// Panics if DB doesn't have a kind of: "mysql", "postgres", or "sqlite3"
//
// Different sql adapters will use different kinds of DSN strings. The strings
// generated here are useful with the following packages:
// MySQL: github.com/go-sql-driver/mysql
// Postgres: github.com/lib/pq
// Sqlite3: code.google.com/p/go-sqlite/go1/sqlite3
//
// This will call DB.DSNSqlite3(useVcsRoot=true) if the kind is Sqlite3.
func (d *DB) DSN() string {
	return d.dsn(true)
}

// DSNnoDB is exactly like DSN except it does not connect directly to the
// database instance, just to the server.
func (d *DB) DSNnoDB() string {
	return d.dsn(false)
}

func (d *DB) dsn(specifyDB bool) string {
	var dsnstr string
	switch d.Kind {
	case "mysql":
		dsnstr = d.mysqlDSN(specifyDB)
	case "postgres":
		dsnstr = d.postgresDSN(specifyDB)
	case "sqlite3":
		dsnstr = d.DSNSqlite3(true)
	default:
		panic("dbm/config: No such database kind: " + d.Kind)
	}

	return dsnstr
}

func (d *DB) mysqlDSN(specifyDB bool) string {
	var dsn bytes.Buffer
	if len(d.User) != 0 {
		dsn.WriteString(d.User)
		if len(d.Pass) != 0 {
			dsn.WriteByte(':')
			dsn.WriteString(d.Pass)
		}
		dsn.WriteByte('@')
	}
	if len(d.Host) != 0 {
		dsn.WriteByte('(')
		dsn.WriteString(d.Host)
		dsn.WriteByte(')')
	}
	dsn.WriteByte('/')
	if specifyDB {
		dsn.WriteString(d.Name)
	}
	return dsn.String()
}

func (d *DB) postgresDSN(specifyDB bool) string {
	var params = make([]string, 0)
	if len(d.User) != 0 {
		params = append(params, fmt.Sprintf("user='%s'", d.User))
	}
	if len(d.Pass) != 0 {
		params = append(params, fmt.Sprintf("password='%s'", d.Pass))
	}
	if len(d.Host) != 0 {
		if strings.HasPrefix(d.Host, "/") {
			params = append(params, fmt.Sprintf("host='%s'", d.Host))
		} else {
			splits := strings.Split(d.Host, ":")

			if len(splits) > 0 && len(splits[0]) != 0 {
				params = append(params, fmt.Sprintf("host='%s'", splits[0]))
			}
			if len(splits) > 1 && len(splits[1]) != 0 {
				params = append(params, fmt.Sprintf("port='%s'", splits[1]))
			}
		}
	}

	if !d.SSL {
		params = append(params, "sslmode=disable")
	} else {
		if !d.SSLSkipVerify {
			params = append(params, "sslmode=verify-full")
		} else {
			params = append(params, "sslmode=require")
		}
	}

	if specifyDB {
		params = append(params, fmt.Sprintf("dbname='%s'", d.Name))
	} else {
		// This is a hack to allow DBCreate
		params = append(params, "dbname='postgres'")
	}
	return strings.Join(params, " ")
}

// DSNSqlite3 creates the filepath for a sqlite3 file.
// If the "name" from the config has the file path separator in it then
// no transformations will be done the path.
//
// If there are no path separators then it will check the useVcsRoot value
// to determine which root directory to use (vcsRoot or cwd). In both cases
// the result will be ROOT/db/{name}.sqlite3
func (d *DB) DSNSqlite3(useVcsRoot bool) string {
	name := d.Name
	if !strings.ContainsRune(name, filepath.Separator) {
		var wd string
		var err error
		if wd, err = os.Getwd(); err != nil {
			panic("Could not get working directory.")
		}

		if useVcsRoot {
			_, vcsRoot, err := paths.FindVCSRoot(wd)
			if err != nil || len(vcsRoot) == 0 {
				panic(fmt.Sprintln("Could not find vcs root:", err))
			} else {
				name = filepath.Join(vcsRoot, DATA_DIR, name)
			}
		} else {
			name = filepath.Join(wd, DATA_DIR, name)
		}
	}
	if len(filepath.Ext(name)) == 0 {
		name += ".sqlite3"
	}
	return name
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
