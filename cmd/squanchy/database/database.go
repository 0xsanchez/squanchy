package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedded embed.FS

// Establishes a new database connection
func Connection(path, mode string) (*sql.DB, error) {
	// Checking for existance
	if _, err := os.Stat(path); err != nil {
		// Creating the database here if the path wasn't specified
		if path == "" || path == "./squanchy.db" {
			fmt.Print("Database file not specified(-d/--database)\nPress <Enter> to create one here: ")
			input := ""
			fmt.Scanln(&input)
			path = "./squanchy.db"
		}
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("error creatine the database %s", err)
		}
		file.Close()
	}
	// Opening the database
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_journal_mode=%s&_fk=true&cache=shared", path, mode))
	if err != nil {
		return nil, fmt.Errorf("error opening the database %s", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	// Using the embbeded migrations
	goose.SetBaseFS(embedded)
	// Setting the dialect
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("error setting the sql dialect %s", err)
	}
	// Applyting migrations to the database
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("error applying migrations %s", err)
	}
	// Pinging the database
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging the database %s", err)
	}
	return db, nil
}
