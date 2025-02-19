package main

import (
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", "migrations", "directory with migration files")
)

func main() {
	flags.Parse(os.Args[1:])

	args := flags.Args()
	if len(args) < 2 {
		flags.Usage()
		return
	}

	dbstring := args[0]
	command := args[1]

	db, err := goose.OpenDBWithDriver("pgx", dbstring)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	if err := goose.Run(command, db, *dir, args[2:]...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
