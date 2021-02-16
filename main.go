package main

import (
	//"restapi/platform/initialize"
	//"database/sql"
	"github.com/jmoiron/sqlx"
	"fmt"
	"log"
	"net/http"
	"restapi/platform/data"
	//"io/ioutil"
	"encoding/json"

	_ "github.com/lib/pq"
)

//Env Create a custom struct which holds a connection pool
type Env struct {
	DB *sqlx.DB
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
	http.HandleFunc("/post", env.PostItems)
	http.ListenAndServe(":3000", nil)
}

// PostItems asdasd
func (env *Env) PostItems(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var feed data.Item
	
		err := json.NewDecoder(r.Body).Decode(&feed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
				
		err = data.PostAll(env.DB, &feed)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		fmt.Fprintf(w, "Data: %+v", feed)
	}
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
		fmt.Fprintf(w, "%s", bk.Post)
	}
}
