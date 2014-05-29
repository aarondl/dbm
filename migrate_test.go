package main

import (
	"database/sql"
	. "testing"
)

type fakeTx struct {
	cmds []string
}

func makeFakeTx() *fakeTx {
	return &fakeTx{
		make([]string, 0),
	}
}

func (f *fakeTx) Exec(cmd string, args ...interface{}) (sql.Result, error) {
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
		tx := makeFakeTx()
		runMigrationPart(tx, []byte(test.Part))

		if len(tx.cmds) != len(test.Expect) {
			t.Errorf("Test failed: %#v", test.Part)
			t.Errorf("Expect: %#v\n", test.Expect)
			t.Errorf("Result: %#v\n", tx.cmds)
		}

		for i := 0; i < len(tx.cmds); i++ {
			if tx.cmds[i] != test.Expect[i] {
				t.Errorf("Test failed: %#v", test.Part)
				t.Errorf("Expect: %#v\n", test.Expect[i])
				t.Errorf("Result: %#v\n", tx.cmds[i])
			}
		}
	}
}
