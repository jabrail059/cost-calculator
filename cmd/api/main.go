package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/handlers"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/middleware"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func main() {
	connStr := "user=postgres password=78552306 dbname=proddb sslmode=disable"
	exec, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	storage.SetDB(exec)
	if err := exec.Ping(); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.Use(middleware.CorsMiddleware)

	router.HandleFunc("/upload", handlers.UploadHandler).Methods("POST")

	router.HandleFunc("/orders", handlers.GetOrdersHandler).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}", handlers.GetOrderByIdHandler).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}/cost", handlers.GetOrderCostHandler).Methods("GET")
	router.HandleFunc("/orders/{id:[0-9]+}/changes", handlers.GetOrderChangesHandler).Methods("GET")

	router.HandleFunc("/api/calculate", handlers.CalculateFromFilesHandler).Methods("POST")

	router.HandleFunc("/orders/{id}/boms", handlers.GetOrderBOMsHandler).Methods("GET")
	router.HandleFunc("/orders/{id}/labor", handlers.GetOrderLaborHandler).Methods("GET")
	router.HandleFunc("/orders/{id}/overhead", handlers.GetOrderOverheadHandler).Methods("GET")

	router.HandleFunc("/reports/generate", handlers.GenerateReportHandler).Methods("POST")

	router.HandleFunc("/mocks/orders", handlers.MockOrdersHandler)
	router.HandleFunc("/mocks/orders/{id:[0-9]+}/cost", handlers.MockOrderCostHandler)

	log.Println("Server is listening...")

	log.Fatal(http.ListenAndServe(":8080", router))
}
