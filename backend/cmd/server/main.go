package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jamespsullivan/pennywise/internal/db"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrate()
		return
	}

	fmt.Println("pennywise server")
}

func runMigrate() {
	dbPath := dbPathFromEnv()

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		log.Fatal(err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = database.Close() }()

	if err := db.Migrate(database); err != nil {
		log.Fatal(err)
	}

	fmt.Println("migrations applied successfully")
}

func dbPathFromEnv() string {
	path := os.Getenv("PENNYWISE_DB_PATH")
	if path == "" {
		path = "./data/pennywise.db"
	}
	return filepath.Clean(path)
}
