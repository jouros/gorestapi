package data

import (
	"database/sql"
	//"fmt"
	"log"
	//"net/http"

	
)

func OpenData() (*sql.DB) {
	db, err := sql.Open("postgres", "postgres://admin:admin123@localhost/dbtest?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	// if could not sql.Open, maybe db.Ping is ok = db is there but something wrong with credentials
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
