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
	"gopkg.in/validator.v2"

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

	// Close db when program exits
	defer db.Close()

	// Create an instance of Env containing the connection pool.
	env := &Env{DB: db}

	// Use env.booksIndex as the handler function for the /books route and env.PostItems for /posts route.
	http.HandleFunc("/data", env.booksIndex)
	http.HandleFunc("/post", env.PostItems)
	http.ListenAndServe(":3000", nil)
}

// PostItems handler
func (env *Env) PostItems(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var feed data.Item // Struct Item is defined in data package
	
		err := json.NewDecoder(r.Body).Decode(&feed) // Unmarshall JSON request body into feed
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// JSON POST input validator, if not valid, return 500 error
		if errs := validator.Validate(feed); errs != nil {
			log.Println("Validate error")
			http.Error(w, http.StatusText(500), 500)
		} else {
			err = data.PostAll(env.DB, &feed) // Pass db connection pool and Item struct to PostAll
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}
		// Return supplied data as a response to client
		fmt.Fprintf(w, "Data: %+v", feed)
	}
}


// Define booksIndex as a method on Env.
func (env *Env) booksIndex(w http.ResponseWriter, r *http.Request) {
	// Pass db connection pool to AllItems
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