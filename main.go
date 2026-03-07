package main

import (
	"database/sql"
	"log"
)

var db *sql.DB

func main() {
	exec, err := sql.Open("postgres", "user=postgres password=78552306 dbname=proddb sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	db = exec
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}
