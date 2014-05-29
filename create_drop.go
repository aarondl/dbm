package main

import (
	"github.com/aarondl/dbm/config"
)

func createDatabase(args []string) {
	engine, err := NewEngine(config.Current)
	if err != nil {
		exitLn("Error getting handle to db:", err)
	}
	if err = engine.CreateDB(); err != nil {
		exitLn("Error creating db:", err)
	}
	trackdbHelper(args, engine)
}

func trackdb(args []string) {
	engine, err := NewEngine(config.Current)
	if err != nil {
		exitLn("Error getting handle to db:", err)
	}
	trackdbHelper(args, engine)
}

func trackdbHelper(args []string, engine SqlEngine) {
	if err := engine.Open(); err != nil {
		exitLn("Error opening to db:", err)
	}
	defer engine.Close()
	if err := engine.CreateMigrationsTable(); err != nil {
		exitLn("Error creating migrations table:", err)
	}
}

func dropDatabase(args []string) {
	engine, err := NewEngine(config.Current)
	if err != nil {
		exitLn("Error getting handle to db:", err)
	}
	if err = engine.DropDB(); err != nil {
		exitLn("Error dropping db:", err)
	}
}
