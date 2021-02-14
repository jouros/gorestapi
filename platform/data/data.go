package data

import (
	"database/sql"
)

type Item struct {
	Title string `json:"title"`
	Post  string `json:"post"`
}

func AllItems(db *sql.DB) ([]Item, error) {
    rows, err := db.Query("SELECT post FROM posts")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var bks []Item

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
