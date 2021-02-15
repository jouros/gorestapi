package main

import (
	//"restapi/platform/initialize"
	"database/sql"
	"restapi/platform/data"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

//Env Create a custom struct which holds a connection pool
type Env struct {
	DB *sql.DB
}

func main() {

	// Initialize db
	// initialize.SyncSchemas()

	// Initialise the connection pool.
	db := data.OpenData()

	defer db.Close()

	// Create an instance of Env containing the connection pool.
	env := &Env{DB: db}

	// Use env.booksIndex as the handler function for the /books route.
	http.HandleFunc("/data", env.booksIndex)
	http.ListenAndServe(":3000", nil)
}

// Define booksIndex as a method on Env.
func (env *Env) booksIndex(w http.ResponseWriter, r *http.Request) {
	// We can now access the connection pool directly in our handlers.
	bks, err := data.AllItems(env.DB)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, bk := range bks {
		fmt.Fprintf(w, "%s, %s", bk.Title, bk.Post)
	}
}
