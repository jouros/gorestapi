package data

import (
	//"database/sql"
	"fmt"
	//"github.com/doug-martin/goqu"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	

	//"reflect"
	"github.com/oleiade/reflections"

	//"encoding/json"
	//"fmt"
	"log"
	//"github.com/lib/pq"
)

// Item Data structure
type Item struct {
	Title string `json:"title" validate:"nonzero,regexp=^[a-zA-Z0-9]*$"`
	Post  string `json:"post" validate:"nonzero"`
}

// Testgogu is just testing for fun
func Testgogu(input *Item) {
	// Just testing goqu for fun
	ds := goqu.Insert("posts").Rows(input)
	insertSQL, args, _ := ds.ToSQL()
	log.Println("gogu is here: ", insertSQL, args)
}

// GetValues is func to read values from Item struct
func GetValues(input *Item) (string, string) {

	// Extract Struct field 'Title'
	fieldsToExtract := []string{"Title"}

	var value1 string

	for _, fieldName := range fieldsToExtract {
    	output1, _ := reflections.GetField(input, fieldName)
		value1 = fmt.Sprintf("%v", output1)
	}
	
	// Extract Struct field 'Post'
	fieldsToExtract2 := []string{"Post"}

	var value2 string

	for _, fieldName := range fieldsToExtract2 {
		output2, _ := reflections.GetField(input, fieldName)
		value2 = fmt.Sprintf("%v", output2)
	}
	return value1, value2
}

// PostAll db connection and pointer to Item Struct
func PostAll(db *sqlx.DB, input *Item) error {
	
	Testgogu(input) // Just for fun

	value1, value2 := GetValues(input) // GetValues from Item struct

	// Format sql 
	ins := "INSERT INTO posts (post) VALUES "
	lauseke := fmt.Sprintf(`%v ('{"%v": "%v"}')`, ins, value1, value2)
	log.Println(lauseke) // Print sql

	_, err := db.Exec(lauseke) // Execute sql 
		if err != nil {
			log.Fatal(err)
		}

	return err
}

// AllItems to read all posted data
func AllItems(db *sqlx.DB) ([]Item, error) {
	
	log.Println("AllItems is here")
	
	rows, err := db.Query("SELECT post FROM posts")
	
	if err != nil {
		return nil, err
	
	}
	
	defer rows.Close()

	var bks []Item // bks = Item Struct

	for rows.Next() {
		var bk Item

		err := rows.Scan(&bk.Post)
		if err != nil {
			return nil, err
		}

		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bks, nil
}