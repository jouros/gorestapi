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

// PostAll db connection and pointer to Item Struct
func PostAll(db *sqlx.DB, input *Item) error {
	
	// Just testing goqu for fun
	ds := goqu.Insert("posts").Rows(input)
	insertSQL, args, _ := ds.ToSQL()
	log.Println("gogu: ", insertSQL, args)

	// Extract Struct field 'Title'
	fieldsToExtract := []string{"Title"}

	for _, fieldName := range fieldsToExtract {
    value1, _ := reflections.GetField(input, fieldName)
	
	// Extract Struct field 'Post'
	fieldsToExtract := []string{"Post"}

	for _, fieldName := range fieldsToExtract {
	value2, _ := reflections.GetField(input, fieldName)
	
	//fmt.Println(value)

	// Format sql 
	ins := "INSERT INTO posts (post) VALUES "
	lauseke := fmt.Sprintf(`%v ('{"%v": "%v"}')`, ins, value1, value2)
	log.Println(lauseke) // Print sql

	_, err := db.Exec(lauseke) // Execute sql 
		if err != nil {
			log.Fatal(err)
		}
	}} // Double For range loop in fieldsToExtract
	
	//fmt.Printf(" PostAll %+v", input)
	return nil
}

// AllItems sdasdasd
func AllItems(db *sqlx.DB) ([]Item, error) {
	rows, err := db.Query("SELECT post FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bks []Item // bks = Item Struct

	for rows.Next() {
		var bk Item

		err := rows.Scan(&bk.Title)
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