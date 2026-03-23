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
	router.HandleFunc("/orders", GetOrdersHandler).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}", GetOrderByIdHandler).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}/cost", GetOrderCostHandler).Methods("GET")
	router.HandleFunc("/api/calculate", CalculateFromFilesHandler).Methods("POST")

	router.HandleFunc("/mocks/orders", MockOrdersHandler)
	router.HandleFunc("/mocks/orders/{id:[0-9]+}/cost", MockOrderCostHandler)

	handler := corsMiddleware(router)
	log.Println("Server is listening...")
	log.Fatal(http.ListenAndServe(":3000", handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
