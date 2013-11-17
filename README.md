Database Manager (dbm)
======

dbm is a command line utility to manage a database via simple migration files.
Although dbm has no advanced generation capabilities and is not as database
agnostic as one might like it should do for a small project to have a little
more structure over how databases are kept up to date inside a team.

## Get Started

Download the package

```bash
go get github.com/aarondl/dbm
```

Create a new configuration, the database, and a new migration
```bash
dbm init
vim db/config # Edit the configuration.
dbm create
dbm new create products
```
(Note the create step will fail with MySQL unless you've created a user with priveleges beforehand or you're using the root user WHICH YOU SHOULD NOT DO)

Edit the new migration:
```
vim db/migrate/20131117212137_create_products.sql
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
 -isroot=false: Set the current working dir as root if set true, otherwise find the first vcs root and use that.
 -v=false: Controls verbose output.
```
