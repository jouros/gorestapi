package data

import (
	//"database/sql"
	"github.com/jmoiron/sqlx"
	//"fmt"
	"log"
	//"net/http"

	
)

// OpenData connection
func OpenData() (*sqlx.DB) {
	db, err := sqlx.Open("postgres", "postgres://admin:admin123@postgres/dbtest?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	// Check Ping also
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}