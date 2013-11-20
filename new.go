package main

import (
	"fmt"
	"github.com/aarondl/dbm/config"
	"github.com/aarondl/paths"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

const (
	_MIG_DIR       = "migrate"
	_MIG_SEPERATOR = "!========================!"
)

// Constants for creation of migrations.
const (
	defMigrationName = "new_migration"
	timeLayout       = "20060102150405_"
)

// migLayout defines the layout for an SQL file
const migLayout = `
` +
	`// Up migration code goes here
` +
	_MIG_SEPERATOR +
	`
// Down migration code goes here
`

var (
	rgxMigrate        = regexp.MustCompile(`^([a-z]|[a-z][a-z_]*[a-z])$`)
	migrationTemplate = template.Must(template.New("mig").Parse(migLayout))
)

func newMigration(args []string) {
	var err error

	migName := "new_migration"
	if len(args) > 0 {
		migName = strings.Join(args, "_")
		if !rgxMigrate.MatchString(migName) {
			exitLn("Invalid migration name:", migName)
		}
	}

	migName = time.Now().Format(timeLayout) + migName + ".sql"

	dir := filepath.Join(workingDir, _DATA_DIR, _MIG_DIR)
	file := filepath.Join(dir, migName)

	if _, err := paths.EnsureDirectory(dir); err != nil {
		exitLn("Could not create migration directory:", err)
	}

	var f *os.File
	if f, err = os.Create(file); err != nil {
		exitLn("Could not create migration file:", err)
	}

	if err = migrationTemplate.Execute(f, config.Current.Name); err != nil {
		exitLn("Error writing to file:", err)
	}

	fmt.Println("Create:", migName)
}
