package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/handlers"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/middleware"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT-ключ не задан")
	}
	connStr := "user=postgres password=78552306 dbname=proddb sslmode=disable"
	exec, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	storage.SetDB(exec)
	if err := exec.Ping(); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.CorsMiddleware)

	r.With(middleware.AuthMiddleware).Post("/upload", handlers.UploadHandler)
	r.Get("/orders", handlers.GetOrdersHandler)
	r.With(middleware.AuthMiddleware).Post("/orders", handlers.CreateOrderHandler)

	r.Get("/orders/{id}", handlers.GetOrderByIdHandler)
	r.Get("/orders/{id}/cost/{method}", handlers.GetOrderCostHandler)
	r.Get("/orders/{id}/changes", handlers.GetOrderChangesHandler)

	r.With(middleware.AuthMiddleware).Post("/api/calculate", handlers.CalculateFromFilesHandler)

	r.Get("/orders/{id}/boms", handlers.GetOrderBOMsHandler)
	r.Get("/orders/{id}/labor", handlers.GetOrderLaborHandler)
	r.Get("/orders/{id}/overhead", handlers.GetOrderOverheadHandler)

	r.With(middleware.AuthMiddleware).Post("/reports/generate", handlers.GenerateReportHandler)
	r.With(middleware.AuthMiddleware).Get("/reports/{id}/excel", handlers.DownloadReportExcelHandler)

	r.Post("/api/auth/register", handlers.RegisterHandler)
	r.Post("/api/auth/login", handlers.LoginHandler)

	r.Get("/mocks/orders", handlers.MockOrdersHandler)
	r.Get("/mocks/orders/{id}/cost", handlers.MockOrderCostHandler)

	r.Handle("/*", http.FileServer(http.Dir("./web")))
	log.Println("Server is listening...")

	log.Fatal(http.ListenAndServe(":8080", r))
}
