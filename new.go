package main

import (
	"fmt"
	"github.com/aarondl/paths"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

const (
	_MIG_DIR = "migrate"
)

const (
	defMigrationName = "new_migration"
	timeLayout       = "20060102150405_"
)

var (
	rgxMigrate        = regexp.MustCompile(`^([a-z]|[a-z][a-z_]*[a-z])$`)
	migrationTemplate = template.Must(
		template.New("mig").Parse("use {{.}};\n\n// SQL Migration code here"),
	)
)

func newMigration(args []string) {
	var err error

	migName := "new_migration"
	if len(args) > 0 {
		migName = strings.Join(args, "_")
		if !rgxMigrate.MatchString(migName) {
			fmt.Println("Invalid migration name:", migName)
			os.Exit(1)
		}
	}

	migName = time.Now().Format(timeLayout) + migName + ".sql"

	dir := filepath.Join(workingDir, _DATA_DIR, _MIG_DIR)
	file := filepath.Join(dir, migName)

	if _, err := paths.EnsureDirectory(dir); err != nil {
		fmt.Println("Could not create migration directory:", err)
		os.Exit(1)
	}

	var f *os.File
	if f, err = os.Create(file); err != nil {
		fmt.Println("Could not create migration file:", err)
		os.Exit(1)
	}

	if err = migrationTemplate.Execute(f, config[*environ].Name); err != nil {
		fmt.Println("Error writing to file:", err)
		os.Exit(1)
	}

	fmt.Println("Create:", migName)
}
