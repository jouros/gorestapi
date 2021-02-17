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
	Title string `json:"title"`
	Post  string `json:"post"`
}

// PostAll sdsad
func PostAll(db *sqlx.DB, input *Item) error {
	
	ds := goqu.Insert("posts").Rows(input)
	insertSQL, args, _ := ds.ToSQL()
	log.Println("gogu: ", insertSQL, args)

	fieldsToExtract := []string{"Title"}

	for _, fieldName := range fieldsToExtract {
    value1, _ := reflections.GetField(input, fieldName)
	

	fieldsToExtract := []string{"Post"}
	for _, fieldName := range fieldsToExtract {
	value2, _ := reflections.GetField(input, fieldName)
    
	//fmt.Println(value)

	ins := "INSERT INTO posts (post) VALUES "
	lauseke := fmt.Sprintf(`%v ('{"%v": "%v"}')`, ins, value1, value2)
	log.Println(lauseke)

	_, err := db.Exec(lauseke) 
		if err != nil {
			log.Fatal(err)
		}
	}}
	
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

	var bks []Item

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
