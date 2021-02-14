package data

import (
	"database/sql"
	"dbtest/data"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

//Env Create a custom struct which holds a connection pool
type Env struct {
	DB *sql.DB
}

func OpenData() (*sql.DB) {
	db, err := sql.Open("postgres", "postgres://admin:admin123@localhost/dbtest?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	// if could not sql.Open, maybe db.Ping is ok = db is there but something wrong with credentials
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Create an instance of Env containing the connection pool.
	env := &Env{DB: db}

	return env
}
