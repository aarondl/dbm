package main

import (
	"database/sql"
	. "testing"
)

type fakeEngine struct {
	cmds []string
}

func makeEngine() *fakeEngine {
	return &fakeEngine{
		make([]string, 0),
	}
}

func (f *fakeEngine) CreateMigrationsTable() error   { return nil }
func (f *fakeEngine) AddMigration(_ string) error    { return nil }
func (f *fakeEngine) DeleteMigration(_ string) error { return nil }
func (f *fakeEngine) Open() error                    { return nil }
func (f *fakeEngine) Close() error                   { return nil }
func (f *fakeEngine) CreateDB() error                { return nil }
func (f *fakeEngine) DropDB() error                  { return nil }

func (f *fakeEngine) Query(_ string, _ ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (f *fakeEngine) Use() error { return nil }

func (f *fakeEngine) Exec(cmd string, args ...interface{}) (sql.Result, error) {
	f.cmds = append(f.cmds, cmd)
	return nil, nil
}

var partTests = []struct {
	Part   string
	Expect []string
}{
	{
		"a /* b;\n */; c;",
		[]string{"a /* b;\n */;", " c;"},
	},
	{
		"a--b;\nc;d;",
		[]string{"a--b;\nc;", "d;"},
	},
	{
		"a#b;\nc;d;",
		[]string{"a#b;\nc;", "d;"},
	},
	{
		"a'/*--;#`\"';b;",
		[]string{"a'/*--;#`\"';", "b;"},
	},
	{
		"a\"/*--;#`'\";b;",
		[]string{"a\"/*--;#`'\";", "b;"},
	},
	{
		"a`/*--;#'\"`;b;",
		[]string{"a`/*--;#'\"`;", "b;"},
	},
}

func Test_RunMigrationPart(t *T) {
	for _, test := range partTests {
		eng := makeEngine()
		runMigrationPart(eng, []byte(test.Part))

		if len(eng.cmds) != len(test.Expect) {
			t.Errorf("Test failed: %#v", test.Part)
			t.Errorf("Expect: %#v\n", test.Expect)
			t.Errorf("Result: %#v\n", eng.cmds)
		}

		for i := 0; i < len(eng.cmds); i++ {
			if eng.cmds[i] != test.Expect[i] {
				t.Errorf("Test failed: %#v", test.Part)
				t.Errorf("Expect: %#v\n", test.Expect[i])
				t.Errorf("Result: %#v\n", eng.cmds[i])
			}
		}
	}
}
