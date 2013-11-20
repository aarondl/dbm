package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/aarondl/dbm/config"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const errFmtNoMatch = `Error: Migrations are out of sync
Possible problems:
   Migration file starting with "%s" missing OR
   Migration file created after migrations run: %s
`
const errOutOfSync = `Error: Migrations are out of sync
The following migration files are missing:`

var rgxUpDown = regexp.MustCompile(`(?s)(.*?)(?:\s` + _MIG_SEPERATOR + `\s(.*))?`)

func doMigrations(args []string) {
	engine, files, done, err := getMigrationData()
	defer engine.Close()
	if err != nil {
		exitLn("Error getting migration data:", err)
	}

	diff := len(files) - len(done)
	if diff == 0 {
		exitLn("Up to date.")
	}

	step := getStep(args)
	if step == 0 || step > diff {
		step = diff
	}

	ensureRunMigrationsMatch(files, done)

	toMigrate := files[len(done) : len(done)+step]
	fmt.Println("Running", step, "migrations...")

	for i := 0; i < len(toMigrate); i++ {
		migrate(engine, toMigrate[i], false)
	}
}

func doRollback(args []string) {
	engine, files, done, err := getMigrationData()
	defer engine.Close()
	if err != nil {
		exitLn("Error getting migration data:", err)
	}

	if len(done) == 0 {
		exitLn("Nothing to rollback.")
	}

	step := getStep(args)
	if step == 0 {
		step = 1
	}
	if step > len(done) {
		step = len(done)
	}

	ensureRunMigrationsMatch(files, done)

	toRollback := files[len(done)-step : len(done)]
	fmt.Println("Rolling back", step, "migrations...")

	for i := len(toRollback) - 1; i >= 0; i-- {
		migrate(engine, toRollback[i], true)
	}
}

func migrate(engine SqlEngine, migration string, rollback bool) {
	var err error

	shortname := filepath.Base(migration)
	up, down := getMigrationParts(migration, shortname)
	if *verbose {
		fmt.Println("=====================================")
		fmt.Println(shortname)
		fmt.Println("=====================================")
	}

	if rollback {
		if len(down) == 0 {
			exitLn("Tried to rollback migration without down:", shortname)
		}

		runMigrationPart(engine, down)
		err = engine.DeleteMigration(migFormat(migration))
	} else {
		runMigrationPart(engine, up)
		err = engine.AddMigration(migFormat(migration))
	}
	if err != nil {
		exitLn("Error updating migration table:", err)
	}
}

func getMigrationParts(filename, shortname string) ([]byte, []byte) {
	f, err := os.Open(filename)
	if err != nil {
		exitLn("Could not open file:", shortname, "-", err)
	}
	defer f.Close()

	var up, down bytes.Buffer
	var sep = []byte(_MIG_SEPERATOR)
	var doingDown = false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			exitLn("Failed to read file:", shortname, "-", err)
		}

		if bytes.Equal(sep, scanner.Bytes()) {
			doingDown = true
			continue
		}

		if doingDown {
			down.Write(scanner.Bytes())
			down.WriteByte('\n')
		} else {
			up.Write(scanner.Bytes())
			up.WriteByte('\n')
		}
	}

	return up.Bytes(), down.Bytes()
}

func runMigrationPart(engine SqlEngine, part []byte) {
	var quote, dblQuote, backQuote bool

	lastIndex := 0
	for i := 0; i < len(part); i++ {
		switch part[i] {
		case '\'':
			if dblQuote || backQuote {
				break
			}
			quote = !quote
		case '"':
			if quote || backQuote {
				break
			}
			dblQuote = !dblQuote
		case '`':
			if quote || dblQuote {
				break
			}
			backQuote = !backQuote
		case '-':
			if i+1 >= len(part) || part[i+1] != '-' {
				break
			}
			fallthrough
		case '#':
			if quote || dblQuote || backQuote {
				break
			}
			for part[i] != '\n' {
				i++
			}
		case '/':
			if quote || dblQuote || backQuote {
				break
			}
			if i+1 < len(part) && part[i+1] == '*' {
				i += 2
				for i < len(part) && !(part[i-1] == '*' && part[i] == '/') {
					i++
				}
			}
		case ';':
			if quote || dblQuote || backQuote {
				break
			}
			cmd := string(part[lastIndex : i+1])
			if _, err := engine.Exec(cmd); err != nil {
				exitf("Error in statement:\nStmt: %s\nErr: %v\n", cmd, err)
			} else if *verbose {
				fmt.Println(strings.TrimSpace(cmd))
			}
			lastIndex = i + 1
		}
	}
}

func getStep(args []string) int {
	var step int
	if len(args) > 0 {
		if s, err := strconv.ParseInt(args[0], 10, 32); err != nil {
			exitf(`Error: Number of steps "%s" must be numerical\n`, args[0])
		} else {
			step = int(s)
		}
	}
	return step
}

func ensureRunMigrationsMatch(files, done []string) {
	var i, j int
	var filesLeft, doneLeft bool

	for {
		filesLeft, doneLeft = i < len(files), j < len(done)
		if !filesLeft || !doneLeft {
			break
		}

		if migFormat(files[i]) != done[j] {
			exitf(errFmtNoMatch, done[j], filepath.Base(files[i]))
		}
		i, j = i+1, j+1
	}

	if doneLeft {
		fmt.Println(errOutOfSync)
		for _, mig := range done[j:] {
			fmt.Printf("    %s\n", mig)
		}
	}
}

func getMigrationData() (SqlEngine, []string, []string, error) {
	var engine SqlEngine
	var files, done []string
	var err error

	if engine, err = NewEngine(config.Current); err != nil {
		exitLn("Error getting handle to db:", err)
	}

	if err = engine.Open(); err != nil {
		exitLn("Failed to connect to database.", err)
	}

	if files, err = getMigrations(); err != nil {
		return nil, nil, nil, err
	}

	if done, err = getRunMigrations(engine); err != nil {
		return nil, nil, nil, err
	}

	return engine, files, done, nil
}

func getMigrations() ([]string, error) {
	var err error
	paths := make([]string, 0)
	path := filepath.Join(workingDir, _DATA_DIR, _MIG_DIR)
	filepath.Walk(path, func(p string, fi os.FileInfo, e error) error {
		if e != nil {
			err = e
			return e
		}
		if !fi.IsDir() && filepath.Ext(p) == ".sql" {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func getRunMigrations(engine SqlEngine) ([]string, error) {
	paths := make([]string, 0)
	result, err := engine.Query(
		fmt.Sprintf("SELECT migration FROM %s;", _MIG_TABLE_NAME))
	if err != nil {
		return nil, err
	}
	for result.Next() {
		var name string
		if err := result.Scan(&name); err != nil {
			return nil, err
		}
		paths = append(paths, name)
	}
	if err = result.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}

func migFormat(migrationfile string) string {
	return strings.Split(filepath.Base(migrationfile), "_")[0]
}
