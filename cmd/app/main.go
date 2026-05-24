package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/config"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/server"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	storage.SetDB(db)
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	router := server.NewRouter(cfg)
	log.Printf("Server is listening on %s", cfg.ServerAddr)

	if cfg.FrontendAPIAddr != "" && cfg.FrontendAPIAddr != cfg.ServerAddr {
		go func() {
			log.Printf("Frontend API compatibility server is listening on %s", cfg.FrontendAPIAddr)
			if err := http.ListenAndServe(cfg.FrontendAPIAddr, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("Frontend API compatibility server stopped: %v", err)
			}
		}()
	}

	log.Fatal(http.ListenAndServe(cfg.ServerAddr, router))
}
