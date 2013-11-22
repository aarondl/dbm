Database Manager (dbm)
======

dbm is a command line utility to manage a database via simple migration files.
Although dbm has no advanced generation capabilities and is not as database
agnostic as one might like it should do for a small project to have a little
more structure over how databases are kept up to date inside a team.

__Supported Databases:__ MySQL, Sqlite3, Postgres

__Warning:__ When you run the create command a bookkeeping table is created
(tracked_migrations), if you remove this table the tool has no idea what
migrations have been run and chaos will ensue!

## Quick Start

The following commands download the package, create a basic config, create the database,
make a new migration and run it against the database.

```bash
go get github.com/aarondl/dbm  # Download package
dbm init                       # Create basic configuration
vim db/config                  # Edit the configuration.
dbm create                     # Create database (and bookeeping table)
dbm new create products        # Create new migration
vim db/migrate/20131117212137_create_products.sql # Edit the migration
dbm migrate                    # Run all migrations
```
_(Note the create step will fail with MySQL unless you've created a user with
priveleges beforehand or you're using the root user WHICH YOU SHOULD NOT DO)_

If we decide that the last two migrations were no good, we can roll them back:
```bash
dbm rollback -v 2 # Passing -v shows us the sql being run.
```

## Connect from Client application

The config package (github.com/aarondl/dbm/config) allows a Go client to load
the configuration used by dbm and generate DSN strings to connect to a configured
database instance.

Config file:
```toml
[development]
kind = "postgres"
name = "dev"
pass = "supersecretpassword"
```

Client file:

```go
package main

import (
	"github.com/aarondl/config"
	"log"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)

func main() {
	if err := config.Load("development"); err != nil {
		log.Fatalln(err)
	}

	db, err := sql.Open("postgres", config.Current.DSN())
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()
}
```

See the docs of the config package for more details.

## Migration Files

The migration files are very particular. The commands MUST end in a ; for them
to be successfully parsed, as there is actually some basic SQL parsing going on
in order to separate multiple statements (Go's SQL interface does not allow for
multiple statements yet).

Up and Down sections are created inside the migration files by using a special
token on it's own line between the sections. This will be inserted for you when
you use the new command. It is possible to create a migration with no down
method (simply delete the special token). However you will be unable to roll any
migration back that is missing a down section.

Here is a sample migration file:

```sql
CREATE TABLE my_table (
  id INTEGER AUTO_INCREMENT NOT NULL,
  name VARCHAR(255)
);
!========================!
DROP TABLE my_table;
```

## Detailed Usage

```text
dbm command [flags] commandArgs
Commands:
 init                    - Create a basic configuration file.
 new      [name]...      - Create a new named migration.
 migrate  [step]         - Migrate [step] forward, migrate all if no step number given.
 rollback [step]         - Rollback [step] backward, rollback most recent if no step number given.
 create                  - Create the configured database.
 drop                    - Drop the configured database.
Flags:
 -env=development: Set the enviroment to choose from the config file.
 -isroot=false: If true use cwd as root, otherwise find VCS root.
 -v=false: Controls verbose output.
```
