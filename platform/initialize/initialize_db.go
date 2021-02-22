package initialize

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// SyncSchemas initialize db schema
func SyncSchemas() {
	// initialize omadb in K8s 
	m, err := migrate.New(
		"file://platform/initialize/sql",
		"postgres://admin:admin123@postgres:5432/omadb?sslmode=disable")
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
