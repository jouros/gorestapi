package initialize

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func SyncSchemas() {
	// initialize dbtest in localhost
	m, err := migrate.New(
		"file://initialize/sql",
		"postgres://admin:admin123@localhost:5432/dbtest?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to start db schema migrator %v", err)
	}
	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			log.Println("schema_migrations version problem")
		} else {
			log.Fatalf("Unable to migrate up to the latest database schema %v", err)
		}
	}
}
