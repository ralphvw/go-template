package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/ralphvw/go-template/helpers"
)

const (
	migrationsDir = "file://./db/migrations"
)

func InitDb() *sql.DB {
	if err := godotenv.Load(); err != nil {
		helpers.LogAction("Error loading env file")
	}
	var err error

	db, err := sql.Open("postgres", os.Getenv("FCC_DB_URL"))

	if err != nil {
		log.Fatal(err)
	}

	helpers.LogAction("DB Connected successfully")

	if err := applyMigrations(); err != nil {
		log.Fatal(err)
	}

	return db
}

func applyMigrations() error {
	m, err := migrate.New(migrationsDir, os.Getenv("FCC_DB_URL"))
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
