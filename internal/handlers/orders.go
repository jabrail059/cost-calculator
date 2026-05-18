package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := storage.GetOrders()
	if err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func GetOrderByIdHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный идентификатор заказа", http.StatusBadRequest)
		return
	}

	order, err := storage.GetOrderByID(id)
	if err == sql.ErrNoRows {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func GetOrderCostHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный формат идентификатора заказа", http.StatusBadRequest)
		return
	}

	method := chi.URLParam(r, "method")
	costResp, err := storage.CalculateCost(id, method)
	if err != nil {
		http.Error(w, "Не удалось посчитать себестоимость", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costResp)
}

func GetOrderBOMsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный идентификатор заказа", http.StatusBadRequest)
		return
	}

	items, err := storage.GetOrderBOMs(id)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func GetOrderLaborHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный идентификатор заказа", http.StatusBadRequest)
		return
	}

	items, err := storage.GetOrderLabor(id)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func GetOrderOverheadHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный идентификатор заказа", http.StatusBadRequest)
		return
	}

	items, err := storage.GetOrderOverhead(id)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateOrderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusBadRequest)
		return
	}

	exists, err := storage.OrderExists(req.ID)
	if err != nil {
		http.Error(w, "Не удалось получить данные из базы данных", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Заказ с таким id уже существует", http.StatusConflict)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, "Не удалось получить start_date", http.StatusInternalServerError)
		return
	}

	var endDate any
	if req.EndDate == "" {
		endDate = nil
	} else {
		endDate, err = time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			http.Error(w, "Не удалось получить end_date", http.StatusInternalServerError)
			return
		}
	}

	err = storage.CreateOrder(req.ID, startDate, endDate, req.Status)
	if err != nil {
		http.Error(w, "Не удалось создать заказ", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok", "id": req.ID})
}
