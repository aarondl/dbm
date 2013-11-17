package main

func createDatabase(args []string) {
	engine, err := NewEngine(config)
	if err != nil {
		exitLn("Error connecting to db:", err)
	}
	defer engine.Close()
	if err = engine.CreateDB(); err != nil {
		exitLn("Error creating db:", err)
	}
}

func dropDatabase(args []string) {
	engine, err := NewEngine(config)
	if err != nil {
		exitLn("Error conneting to db:", err)
	}
	defer engine.Close()
	if err = engine.DropDB(); err != nil {
		exitLn("Error dropping db:", err)
	}
}
