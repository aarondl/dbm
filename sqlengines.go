package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/aarondl/dbm/config"
	"github.com/aarondl/paths"
	"os"
	"path/filepath"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	_MIG_TABLE_NAME   = "tracked_migrations"
	sqlUseDB          = `use %s;`
	sqlCreateDB       = `CREATE DATABASE IF NOT EXISTS %s;`
	sqlCreateDBPQ     = `CREATE DATABASE %s;`
	sqlAddMig         = `INSERT INTO %s (migration) VALUES (?);`
	sqlAddMigPQ       = `INSERT INTO %s (migration) VALUES ($1);`
	sqlDelMig         = `DELETE FROM %s WHERE migration=?;`
	sqlDelMigPQ       = `DELETE FROM %s WHERE migration=$1;`
	sqlDropDB         = `DROP DATABASE IF EXISTS %s;`
	sqlWipeTrackTable = `DELETE FROM ` + _MIG_TABLE_NAME + `;`
)

const sqlCreateTrackTable = `
CREATE TABLE IF NOT EXISTS ` + _MIG_TABLE_NAME + ` (
	migration varchar(255) NOT NULL
);`

func NewEngine(conf *config.DB) (SqlEngine, error) {
	if len(conf.Name) == 0 {
		return nil, errors.New("dbm: Database must have a name.")
	}

	switch conf.Kind {
	case "mysql":
		return NewMySQL(conf)
	case "postgres":
		return NewPostgres(conf)
	case "sqlite3":
		return NewSqlite3(conf)
	default:
		return nil, fmt.Errorf("dbm: Unknown db engine: %s", conf.Kind)
	}
}

type SqlEngine interface {
	// CreateDB creates the database, does not require Open() first.
	CreateDB() error
	// DropDB drops the database, does not require Open() first.
	DropDB() error

	// Open the connection to the specified database.
	Open() error
	// Close the connection to the database.
	Close() error
	// CreateMigrationsTable adds a tracking table for migrations.
	CreateMigrationsTable() error
	// AddMigration adds a tracking record for a migration.
	AddMigration(mig string) error
	// DeleteMigration removes a tracking record for a migration.
	DeleteMigration(mig string) error
	// Exec executes a statement against the database.
	Exec(stmt string, args ...interface{}) (sql.Result, error)
	// Query executes a query against the database.
	Query(stmt string, args ...interface{}) (*sql.Rows, error)
}

type MySQL struct {
	conf *config.DB
	*sql.DB
}

func NewMySQL(d *config.DB) (*MySQL, error) {
	return &MySQL{conf: d}, nil
}

func (m *MySQL) Open() error {
	var err error
	m.DB, err = sql.Open("mysql", m.conf.DSN())
	return err
}

func (m *MySQL) CreateDB() error {
	var err error
	if m.DB, err = sql.Open("mysql", m.conf.DSNnoDB()); err != nil {
		return err
	}
	defer m.Close()

	if _, err := m.Exec(fmt.Sprintf(sqlCreateDB, m.conf.Name)); err != nil {
		return err
	}

	return nil
}

func (m *MySQL) DropDB() error {
	var err error
	if m.DB, err = sql.Open("mysql", m.conf.DSNnoDB()); err != nil {
		return err
	}
	defer m.Close()

	if _, err = m.Exec(fmt.Sprintf(sqlDropDB, m.conf.Name)); err != nil {
		return err
	}

	return nil
}

func (m *MySQL) CreateMigrationsTable() error {
	return createTrackTable(m)
}

func (m *MySQL) AddMigration(mig string) error {
	return insertTrackTable(m, sqlAddMig, mig)
}

func (m *MySQL) DeleteMigration(mig string) error {
	return deleteTrackTable(m, sqlDelMig, mig)
}

type Postgres struct {
	conf *config.DB
	*sql.DB
}

func NewPostgres(d *config.DB) (*Postgres, error) {
	return &Postgres{conf: d}, nil
}

func (p *Postgres) Open() error {
	var err error
	p.DB, err = sql.Open("postgres", p.conf.DSN())
	return err
}

func (p *Postgres) CreateDB() error {
	var err error
	if p.DB, err = sql.Open("postgres", p.conf.DSNnoDB()); err != nil {
		return err
	}
	defer p.Close()

	if _, err := p.Exec(fmt.Sprintf(sqlCreateDBPQ, p.conf.Name)); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DropDB() error {
	var err error
	if p.DB, err = sql.Open("postgres", p.conf.DSNnoDB()); err != nil {
		return err
	}
	defer p.Close()

	if _, err = p.Exec(fmt.Sprintf(sqlDropDB, p.conf.Name)); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) CreateMigrationsTable() error {
	return createTrackTable(p)
}

func (p *Postgres) AddMigration(mig string) error {
	return insertTrackTable(p, sqlAddMigPQ, mig)
}

func (p *Postgres) DeleteMigration(mig string) error {
	return deleteTrackTable(p, sqlDelMigPQ, mig)
}

type Sqlite3 struct {
	conf *config.DB
	*sql.DB
	path string
}

func NewSqlite3(d *config.DB) (*Sqlite3, error) {
	return &Sqlite3{
		conf: d,
		path: d.DSNSqlite3(!*isRoot),
	}, nil
}

func (s *Sqlite3) Open() error {
	var err error
	s.DB, err = sql.Open("sqlite3", s.path)
	return err
}

func (s *Sqlite3) CreateDB() error {
	dir := filepath.Dir(s.path)
	if _, err := paths.EnsureDirectory(dir); err != nil {
		return err
	}
	return nil
}

func (s *Sqlite3) DropDB() error {
	return os.Remove(s.path)
}

func (s *Sqlite3) CreateMigrationsTable() error {
	return createTrackTable(s)
}

func (s *Sqlite3) AddMigration(mig string) error {
	return insertTrackTable(s, sqlAddMig, mig)
}

func (s *Sqlite3) DeleteMigration(mig string) error {
	return deleteTrackTable(s, sqlDelMig, mig)
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

func insertTrackTable(engine SqlEngine, sql, mig string) error {
	_, err := engine.Exec(fmt.Sprintf(sql, _MIG_TABLE_NAME), mig)
	return err
}

func deleteTrackTable(engine SqlEngine, sql, mig string) error {
	_, err := engine.Exec(fmt.Sprintf(sql, _MIG_TABLE_NAME), mig)
	return err
}
