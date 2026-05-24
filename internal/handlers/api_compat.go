package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

type frontendFile struct {
	Name string `json:"name"`
}

type frontendOrder struct {
	OrderNumber string         `json:"orderNumber"`
	Price       float64        `json:"price"`
	Cost        float64        `json:"cost"`
	Margin      float64        `json:"margin"`
	StartDate   string         `json:"startDate"`
	EndDate     string         `json:"endDate"`
	Status      string         `json:"status"`
	Files       []frontendFile `json:"files"`
}

func GetAPIOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := storage.GetOrders()
	if err != nil {
		writeJSONError(w, "failed to load orders", http.StatusInternalServerError)
		return
	}

	result := make([]frontendOrder, 0, len(orders))
	for _, order := range orders {
		fileNames, err := storage.GetOrderFileNames(order.Id)
		if err != nil {
			writeJSONError(w, "failed to load order files", http.StatusInternalServerError)
			return
		}
		files := make([]frontendFile, 0, len(fileNames))
		for _, name := range fileNames {
			files = append(files, frontendFile{Name: name})
		}

		endDate := ""
		if order.EndDate != nil {
			endDate = order.EndDate.Format("2006-01-02")
		}

		result = append(result, frontendOrder{
			OrderNumber: fmt.Sprintf("%d", order.Id),
			Price:       order.TotalCost,
			Cost:        order.TotalCost,
			Margin:      0,
			StartDate:   order.StartDate.Format("2006-01-02"),
			EndDate:     endDate,
			Status:      order.Status,
			Files:       files,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func CreateAPIOrderHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(handlerConfig.UploadMaxMemory); err != nil {
		writeJSONError(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	id, err := parseFrontendOrderID(r.FormValue("orderNumber"))
	if err != nil {
		writeJSONError(w, "invalid order number", http.StatusBadRequest)
		return
	}

	startDate, err := parseFrontendDate(r.FormValue("startDate"))
	if err != nil {
		writeJSONError(w, "invalid start date", http.StatusBadRequest)
		return
	}

	var endDate any
	if value := r.FormValue("endDate"); value != "" {
		parsed, err := parseFrontendDate(value)
		if err != nil {
			writeJSONError(w, "invalid end date", http.StatusBadRequest)
			return
		}
		endDate = parsed
	}

	status := strings.TrimSpace(r.FormValue("status"))
	if status == "" {
		writeJSONError(w, "order status is required", http.StatusBadRequest)
		return
	}

	exists, err := storage.OrderExists(id)
	if err != nil {
		writeJSONError(w, "failed to read database", http.StatusInternalServerError)
		return
	}
	if exists {
		writeJSONError(w, "order already exists", http.StatusConflict)
		return
	}

	if err := storage.CreateOrder(id, startDate, endDate, status); err != nil {
		writeJSONError(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "id": id})
}

func GetAPIHistoryHandler(w http.ResponseWriter, r *http.Request) {
	logs, err := storage.GetUploadLogs()
	if err != nil {
		writeJSONError(w, "failed to load history", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, 0, len(logs))
	for _, item := range logs {
		result = append(result, map[string]any{
			"date":   item.UploadedAt,
			"user":   item.ChangedBy,
			"role":   "",
			"field":  item.Filetype,
			"oldVal": "",
			"newVal": fmt.Sprintf("Uploaded file for order %d", item.OrderID),
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func GetAPIErrorsHandler(w http.ResponseWriter, r *http.Request) {
	items, err := storage.GetErrors()
	if err != nil {
		writeJSONError(w, "failed to load errors", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"orderNumber": fmt.Sprintf("%d", item.ID),
			"info":        strings.TrimSpace(item.Info + " " + item.RowAndColumn),
			"date":        time.Now(),
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func GenerateAPIReportHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID string `json:"orderId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	reportData := []models.ReportItem{{
		Order: req.OrderID,
		Type:  "total",
		Sum:   0,
	}}
	file, err := generateExcel(reportData)
	if err != nil {
		writeJSONError(w, "failed to create Excel file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
	if err := file.Write(w); err != nil {
		http.Error(w, "failed to create Excel", http.StatusInternalServerError)
	}
}

func parseFrontendOrderID(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("empty order id")
	}
	if id, err := strconv.Atoi(value); err == nil && id > 0 {
		return id, nil
	}

	digits := strings.Builder{}
	for _, r := range value {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	if digits.Len() == 0 {
		return 0, errors.New("order id has no digits")
	}
	id, err := strconv.Atoi(digits.String())
	if err != nil || id <= 0 {
		return 0, errors.New("invalid order id")
	}
	return id, nil
}

func parseFrontendDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ".", "-")
	return time.Parse("2006-01-02", value)
}
