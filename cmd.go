/*
Package dbm does simple database management by creating and running migration
files.

Please see the README.md or use the flag -h for more information.
*/
package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/aarondl/paths"
	"os"
	"path/filepath"
	"regexp"
)

const (
	_CONFIG   = "config.toml"
	_DATA_DIR = "db"
)

const usageDesc = `dbm command [flags] commandArgs
Commands:
    init                    - Create a basic configuration file.
    new      [name]...      - Create a new named migration.
    migrate  [step]         - Migrate [step] forward, migrate all if no step number given.
    rollback [step]         - Rollback [step] backward, rollback all if no step number given.
    create                  - Create the configured database.
    drop                    - Drop the configured database.`

var (
	flagset = flag.NewFlagSet("flags", flag.ExitOnError)
	isRoot  = flagset.Bool("isroot", false,
		`Set the current working dir as root if set true, otherwise find the `+
			`first vcs root and use that.`)
	environ = flagset.String("env", "development",
		`Set the enviroment to choose from the config file.`)
	verbose = flagset.Bool("v", false, "Controls verbose output.")
)

var (
	rgxEnviron = regexp.MustCompile(`^[a-z]+$`)
)

var (
	workingDir string
	configs    map[string]*DbConfig
	config     *DbConfig
)

type DbConfig struct {
	Name string
	Kind string
	Host string
	User string
	Pass string
}

var commands = map[string]func([]string){
	"new":      newMigration,
	"migrate":  doMigrations,
	"rollback": doRollback,
	"create":   createDatabase,
	"drop":     dropDatabase,
	"init":     initialize,
}

func main() {
	// Determine command.
	cmdArgs := os.Args[1:]

	if len(cmdArgs) == 0 {
		printUsage()
	}
	cmd := cmdArgs[0]
	if _, ok := commands[cmd]; !ok {
		printUsage()
	}

	flagset.Parse(cmdArgs[1:]) // Parse flag arguments.
	cmdArgs = flagset.Args()

	// Verify environment
	if !rgxEnviron.MatchString(*environ) {
		fmt.Println("Invalid environment:", *environ,
			" must be lowercase letters only.")
	}

	// Set the working directory.
	setRoot()
	fmt.Println(workingDir)

	// We have to initialize here since we can't load the config if someone
	// wants to create it!
	if cmd == "init" {
		commands[cmd](cmdArgs)
		return
	}

	// Parse the config
	configPath := filepath.Join(workingDir, _DATA_DIR, _CONFIG)
	loadConfig(configPath)

	// Set env
	var ok bool
	if config, ok = configs[*environ]; !ok {
		exitLn("No such configured environment:", *environ)
	}

	// Dispatch the command handler.
	commands[cmd](cmdArgs)
}

func printUsage() {
	fmt.Println(usageDesc)
	fmt.Println("Flags:")
	flagset.VisitAll(func(f *flag.Flag) {
		fmt.Printf("    -%s=%v: %s\n", f.Name, f.DefValue, f.Usage)
	})
	os.Exit(1)
}

func exit(args ...interface{}) {
	fmt.Print(args...)
	os.Exit(1)
}

func exitf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func exitLn(args ...interface{}) {
	fmt.Println(args...)
	os.Exit(1)
}

func setRoot() {
	var err error
	var wd string
	if wd, err = os.Getwd(); err != nil {
		exitLn("Could not get working dir:", err)
	}

	if *isRoot {
		workingDir = wd
		return
	}

	var p string
	if _, p, err = paths.FindVCSRoot(wd); err != nil || len(p) == 0 {
		exitLn("Error: Could not find vcs root.")
	}
	workingDir = p
}

func loadConfig(configPath string) {
	if yes, err := paths.FileExists(configPath); err != nil {
		exitLn("Error: Could not check for config file at:", configPath)
	} else if !yes {
		exitLn("Error: No configuration file at:", configPath)
	}

	var err error
	if _, err = toml.DecodeFile(configPath, &configs); err != nil {
		exitLn("Could not decode config file:", err)
	}
}
