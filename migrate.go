package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/aarondl/paths"
	//"github.com/davecgh/go-spew/spew"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	_CONFIG   = "config.toml"
	_DATA_DIR = "db"
)

const usageDesc = `migrate [globalFlags] command [commandFlags] commandArgs
Commands:
    new [migration name]... - Create a new named migration.
    migrate [step]          - Migrate [step] forward, migrate all if no step number given.
    rollback [step]         - Rollback [step] backward, rollback all if no step number given.
    create                  - Create the configured database.
    drop                    - Drop the configured database.`

var (
	isRoot = flag.Bool("isroot", false,
		`Set the current working dir as root if set true, otherwise find the `+
			`first git root and use that.`)
	environ = flag.String("env", "development",
		`Set the enviroment to choose from the config file.`)
)

var (
	rgxEnviron = regexp.MustCompile(`^[a-z]+$`)
)

var (
	workingDir string
	config     map[string]*DbConfig
)

type DbConfig struct {
	Name     string
	Kind     string
	Url      string
	Username string
	Password string
}

func main() {
	flag.Parse() // Parse flag arguments.

	// Determine command.
	cmdArgs := os.Args[len(os.Args)-flag.NArg():]

	if len(cmdArgs) == 0 {
		fmt.Println(usageDesc)
		fmt.Println("Global Flags:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("    -%s=%v: %s\n", f.Name, f.DefValue, f.Usage)
		})
		return
	}
	cmd := cmdArgs[0]
	cmdArgs = cmdArgs[1:]

	// Verify environment
	if !rgxEnviron.MatchString(*environ) {
		fmt.Println("Invalid environment:", *environ,
			" must be lowercase letters only.")
	}

	// Set the working directory.
	setRoot()

	// Parse the config
	configPath := filepath.Join(workingDir, _DATA_DIR, _CONFIG)
	loadConfig(configPath)

	fmt.Println(workingDir)

	switch cmd {
	case "new":
		newMigration(cmdArgs)
	case "migrate":
	case "rollback":
	case "create":
	case "drop":
	}
}

func setRoot() {
	var err error
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not get working dir:", err)
		os.Exit(1)
	}

	if !*isRoot {
		for p := wd; len(p) != 0; p = paths.WalkUpPath(p) {
			yes, err := paths.DirExists(filepath.Join(p, ".git"))
			if err != nil {
				fmt.Println("Error searching for git root:", err)
				os.Exit(1)
			} else if yes {
				workingDir = p
				break
			}
		}
		if len(workingDir) == 0 {
			fmt.Println("Error: Could not find git root.")
			os.Exit(1)
		}
	} else {
		workingDir = wd
	}
}

func loadConfig(configPath string) {
	if yes, err := paths.FileExists(configPath); err != nil {
		fmt.Println("Error: Could not check for config file at:", configPath)
		os.Exit(1)
	} else if !yes {
		fmt.Println("Error: No configuration file at:", configPath)
		os.Exit(1)
	}

	var err error
	if _, err = toml.DecodeFile(configPath, &config); err != nil {
		fmt.Println("Could not decode config file:", err)
		os.Exit(1)
	}
}
