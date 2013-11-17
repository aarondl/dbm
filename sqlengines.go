package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aarondl/paths"
	"os"
	"path/filepath"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	_ "github.com/go-sql-driver/mysql"
)

const (
	_MIG_TABLE_NAME   = "tracked_migrations"
	sqlUseDB          = `use %s;`
	sqlCreateDB       = `CREATE DATABASE IF NOT EXISTS %s;`
	sqlDropDB         = `DROP DATABASE IF EXISTS %s;`
	sqlWipeTrackTable = `DELETE FROM ` + _MIG_TABLE_NAME + `;`
)

const sqlCreateTrackTable = `
CREATE TABLE IF NOT EXISTS ` + _MIG_TABLE_NAME + ` (
	migration varchar(255) NOT NULL
);`

func NewEngine(conf *DbConfig) (SqlEngine, error) {
	switch conf.Kind {
	case "sqlite3":
		return NewSqlite3(conf)
	case "mysql":
		return NewMySQL(conf)
	default:
		return nil, fmt.Errorf("Unknown db engine:", conf.Kind)
	}
}

type SqlEngine interface {
	CreateDB() error
	DropDB() error
	Close() error
	Use() error
	Exec(stmt string, args ...interface{}) (sql.Result, error)
	Query(stmt string, args ...interface{}) (*sql.Rows, error)
}

type MySQL struct {
	*DbConfig
	*sql.DB
	dsn string
}

func NewMySQL(d *DbConfig) (*MySQL, error) {
	if len(d.Name) == 0 {
		return nil, errors.New("Database must have a name.")
	}

	m := &MySQL{DbConfig: d}
	if err := m.Open(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MySQL) Open() error {
	if len(m.dsn) == 0 {
		m.makeDSN()
	}
	var err error
	m.DB, err = sql.Open("mysql", m.dsn)
	return err
}

func (m *MySQL) CreateDB() error {
	if _, err := m.Exec(fmt.Sprintf(sqlCreateDB, m.Name)); err != nil {
		return err
	}
	if err := m.Use(); err != nil {
		return err
	}
	return createTrackTable(m)
}

func (m *MySQL) Use() error {
	if _, err := m.Exec(fmt.Sprintf(sqlUseDB, m.Name)); err != nil {
		return err
	}
	return nil
}

func (m *MySQL) DropDB() error {
	var err error
	if _, err = m.Exec(fmt.Sprintf(sqlDropDB, m.Name)); err != nil {
		return err
	}
	return nil
}

func (m *MySQL) makeDSN() {
	var dsn bytes.Buffer
	if len(m.Username) != 0 {
		dsn.WriteString(m.Username)
		if len(m.Password) != 0 {
			dsn.WriteByte(':')
			dsn.WriteString(m.Password)
		}
		dsn.WriteByte('@')
	}
	if len(m.Host) != 0 {
		dsn.WriteByte('(')
		dsn.WriteString(m.Host)
		dsn.WriteByte(')')
	}
	dsn.WriteByte('/')
	m.dsn = dsn.String()
}

type Sqlite3 struct {
	*DbConfig
	*sql.DB
	dsn string
}

func NewSqlite3(d *DbConfig) (*Sqlite3, error) {
	if len(d.Name) == 0 {
		return nil, errors.New("Database must have a name.")
	}

	s := &Sqlite3{DbConfig: d}
	if err := s.Open(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Sqlite3) Open() error {
	if len(s.dsn) == 0 {
		s.makeDSN()
	}
	var err error
	s.DB, err = sql.Open("sqlite3", s.dsn)
	return err
}

func (s *Sqlite3) CreateDB() error {
	dir := filepath.Dir(s.dsn)
	if _, err := paths.EnsureDirectory(dir); err != nil {
		return err
	}
	return createTrackTable(s)
}

func (s *Sqlite3) DropDB() error {
	return os.Remove(s.dsn)
}

func (s *Sqlite3) Use() error {
	return nil
}

func (s *Sqlite3) makeDSN() {
	name := s.Name
	if !filepath.IsAbs(s.Name) {
		name = filepath.Join(workingDir, _DATA_DIR, name)
	}
	if len(filepath.Ext(name)) == 0 {
		name += ".sqlite3"
	}
	s.dsn = name
}

func commandReader(first, last string, cmds ...string) *bytes.Buffer {
	var b bytes.Buffer
	if len(first) != 0 {
		b.WriteString(first)
		b.WriteByte('\n')
	}
	for i := 0; i < len(cmds); i++ {
		b.WriteString(cmds[i])
		b.WriteByte('\n')
	}
	if len(last) != 0 {
		b.WriteString(last)
		b.WriteByte('\n')
	}
	return &b
}

func createTrackTable(engine SqlEngine) error {
	var err error
	if _, err = engine.Exec(sqlCreateTrackTable); err != nil {
		return err
	}
	if _, err = engine.Exec(sqlWipeTrackTable); err != nil {
		return err
	}
	return nil
}
