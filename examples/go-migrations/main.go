package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mc2soft/goose"

	// Init DB drivers.
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/godrv"
)

var (
	flags       = flag.NewFlagSet("goose", flag.ExitOnError)
	dir         = flags.String("dir", ".", "directory with migration files")
	missingOnly = flags.Bool("missing-only", false, "for status command - find out only migrations, missing from the current DB")
)

func main() {
	flags.Usage = usage
	flags.Parse(os.Args[1:])

	args := flags.Args()

	if len(args) > 1 && args[0] == "create" {
		if err := goose.Run("create", nil, *dir, *missingOnly, args[1:]...); err != nil {
			log.Fatalf("goose run: %v", err)
		}
		return
	}

	if len(args) < 3 {
		flags.Usage()
		return
	}

	if args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	driver, dbstring, command := args[0], args[1], args[2]

	switch driver {
	case "postgres", "mysql", "sqlite3", "redshift":
		if err := goose.SetDialect(driver); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("%q driver not supported\n", driver)
	}

	switch dbstring {
	case "":
		log.Fatalf("-dbstring=%q not supported\n", dbstring)
	default:
	}

	if driver == "redshift" {
		driver = "postgres"
	}

	db, err := sql.Open(driver, dbstring)
	if err != nil {
		log.Fatalf("-dbstring=%q: %v\n", dbstring, err)
	}

	arguments := []string{}
	if len(args) > 3 {
		arguments = append(arguments, args[3:]...)
	}

	if err := goose.Run(command, db, *dir, *missingOnly, arguments...); err != nil {
		log.Fatalf("goose run: %v", err)
	}
}

func usage() {
	fmt.Print(usagePrefix)
	flags.PrintDefaults()
	fmt.Print(usageCommands)
}

var (
	usagePrefix = `Usage: goose [OPTIONS] DRIVER DBSTRING COMMAND

Drivers:
    postgres
    mysql
    sqlite3
    redshift

Examples:
    goose sqlite3 ./foo.db status
    goose sqlite3 ./foo.db create init sql
    goose sqlite3 ./foo.db create add_some_column sql
    goose sqlite3 ./foo.db create fetch_user_data go
    goose sqlite3 ./foo.db up

    goose postgres "user=postgres dbname=postgres sslmode=disable" status
    goose mysql "user:password@/dbname" status
    goose redshift "postgres://user:password@qwerty.us-east-1.redshift.amazonaws.com:5439/db" status

Options:
    -dir string
        directory with migration files (default ".")
    -missing-only
        for status command - find out only migrations, missing from the current DB
`

	usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate up by a single version
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    status               Dump the migration status for the current DB. Use [-missing-only] option to find out only migrations, missing from the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with next version
`
)
