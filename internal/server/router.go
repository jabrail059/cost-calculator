package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/config"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/handlers"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/middleware"
)

func NewRouter(cfg config.Config) http.Handler {
	handlers.Configure(handlers.Config{
		OneCURL:         cfg.OneCURL,
		OneCTimeout:     cfg.OneCTimeout,
		UploadMaxMemory: cfg.UploadMaxMemory,
	})

	r := chi.NewRouter()
	r.Use(middleware.CorsMiddleware)

	r.With(middleware.AuthMiddleware).Post("/upload", handlers.UploadHandler)
	r.Get("/orders", handlers.GetOrdersHandler)
	r.Post("/orders", handlers.CreateOrderHandler)

	r.Get("/orders/{id}", handlers.GetOrderByIdHandler)
	r.Get("/orders/{id}/cost/{method}", handlers.GetOrderCostHandler)
	r.Get("/orders/{id}/changes", handlers.GetOrderChangesHandler)

	r.Post("/api/calculate", handlers.CalculateFromFilesHandler)
	r.Get("/api/orders", handlers.GetAPIOrdersHandler)
	r.With(middleware.AuthMiddleware).Post("/api/orders", handlers.CreateAPIOrderHandler)
	r.Get("/api/history", handlers.GetAPIHistoryHandler)
	r.Get("/api/errors", handlers.GetAPIErrorsHandler)
	r.Post("/api/report/generate", handlers.GenerateAPIReportHandler)
	r.Post("/api/generate-excel", handlers.GenerateExcelFromOneCHandler)
	r.Get("/orders/{id}/boms", handlers.GetOrderBOMsHandler)
	r.Get("/orders/{id}/labor", handlers.GetOrderLaborHandler)
	r.Get("/orders/{id}/overhead", handlers.GetOrderOverheadHandler)

	r.Post("/reports/generate", handlers.GenerateReportHandler)
	r.Get("/reports/{id}/excel", handlers.DownloadReportExcelHandler)

	r.Post("/api/auth/register", handlers.RegisterHandler)
	r.Post("/api/auth/login", handlers.LoginHandler)
	r.Post("/api/register", handlers.RegisterHandler)
	r.Post("/api/login", handlers.LoginHandler)

	r.Get("/mocks/orders", handlers.MockOrdersHandler)
	r.Get("/mocks/orders/{id}/cost", handlers.MockOrderCostHandler)
	r.Post("/mock/1c/report", handlers.MockOneCReportHandler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui1-signUp.html", http.StatusFound)
	})
	r.Handle("/*", webHandler(cfg.WebDir))
	return r
}
