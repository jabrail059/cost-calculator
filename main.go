package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
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
	router := mux.NewRouter()

	router.HandleFunc("/upload", UploadHandler)
	router.HandleFunc("/mocks/orders", MockOrdersHandler)
	router.HandleFunc("/mocks/orders/{id:[0-9]+}/cost", MockOrderCostHandler)

	log.Println("Server is listening...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
