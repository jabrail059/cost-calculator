package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)

	exec, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	db = exec
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	if err := runMigrations(db); err != nil {
		log.Fatalf("Ошибка при выполнении миграций: %v", err)
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

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
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
