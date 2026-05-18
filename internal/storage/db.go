package storage

import "database/sql"

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func DB() *sql.DB {
	return db
}
