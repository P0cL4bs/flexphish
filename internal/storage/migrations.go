package storage

import (
	"database/sql"
	"embed"
	"flexphish/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(dbPath string) error {

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Log.Error("Failed to open database")
		return err
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		logger.Log.Error("Failed to create sqlite driver")
		return err
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		logger.Log.Error("Failed to load migrations")
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		"sqlite3",
		driver,
	)
	if err != nil {
		logger.Log.Error("Failed to initialize migrate instance")
		return err
	}

	err = m.Up()

	if err == migrate.ErrNoChange {
		return nil
	}

	if err != nil {
		logger.Log.Error("Migration failed")
		return err
	}

	logger.Log.Info("[+] Database migrations applied")

	return nil
}
