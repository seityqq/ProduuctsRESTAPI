package migrate

import (
	_ "database/sql"
	_ "fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	_ "net/http"
)

func RunMigrations(dsn string) {
	m, err := migrate.New(
		"file://migrations", dsn)
	if err != nil {
		log.Fatalf("Error creating migrator: %v", err)
	}

	version, dirty, _ := m.Version()
	if dirty {
		log.Println("Database in 'dirty' state")
	} else {
		log.Printf("Current migration version: %d", version)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error applying migration\n\n: %v", err)
	}

	newVersion, _, _ := m.Version()
	log.Printf("Migrations have been applied. New version: %d", newVersion)
}
