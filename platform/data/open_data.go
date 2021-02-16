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
	db, err := sqlx.Open("postgres", "postgres://admin:admin123@localhost/dbtest?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	// if could not sql.Open, maybe db.Ping is ok = db is there but something wrong with credentials
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
