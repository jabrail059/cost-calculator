package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/parser"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/utils"
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

	if err := saveFrontendOrderFiles(r, id); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "id": id})
}

func saveFrontendOrderFiles(r *http.Request, orderID int) error {
	type uploadFile struct {
		field string
		kind  string
	}
	files := []uploadFile{
		{field: "file_0", kind: "boms"},
		{field: "file_1", kind: "labor"},
		{field: "file_2", kind: "overhead"},
	}

	changedBy := "frontend"
	if user := strings.TrimSpace(r.Header.Get("X-User")); user != "" {
		changedBy = user
	}

	for _, fileInfo := range files {
		file, _, err := r.FormFile(fileInfo.field)
		if errors.Is(err, http.ErrMissingFile) {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to read %s file: %w", fileInfo.kind, err)
		}

		temp, err := os.CreateTemp("", "*")
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tempName := temp.Name()

		_, copyErr := io.Copy(temp, file)
		closeErr := temp.Close()
		file.Close()
		defer os.Remove(tempName)
		if copyErr != nil {
			return fmt.Errorf("failed to copy %s file: %w", fileInfo.kind, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close %s file: %w", fileInfo.kind, closeErr)
		}

		if err := saveParsedOrderFile(tempName, fileInfo.kind, orderID, changedBy); err != nil {
			return err
		}
	}

	return nil
}

func saveParsedOrderFile(path string, kind string, orderID int, changedBy string) error {
	switch kind {
	case "boms":
		items, err := parser.ParseBOM(path)
		if err != nil {
			return err
		}
		if err := ensureOrderID(orderID, utils.ExtractOrderIds(items)); err != nil {
			return err
		}
		return DoTransactions(func(tx *sql.Tx) error {
			if err := storage.InsertBOMItems(tx, items); err != nil {
				return err
			}
			return storage.SaveUploadLog(tx, []int{orderID}, kind, changedBy)
		})
	case "labor":
		items, err := parser.ParseLabor(path)
		if err != nil {
			return err
		}
		if err := ensureOrderID(orderID, utils.ExtractOrderIds(items)); err != nil {
			return err
		}
		return DoTransactions(func(tx *sql.Tx) error {
			if err := storage.InsertLaborItems(tx, items); err != nil {
				return err
			}
			return storage.SaveUploadLog(tx, []int{orderID}, kind, changedBy)
		})
	case "overhead":
		items, err := parser.ParseOverhead(path)
		if err != nil {
			return err
		}
		for i := range items {
			if items[i].OrderID == nil {
				id := orderID
				items[i].OrderID = &id
			}
		}
		return DoTransactions(func(tx *sql.Tx) error {
			if err := storage.InsertOverheadItems(tx, items); err != nil {
				return err
			}
			return storage.SaveUploadLog(tx, []int{orderID}, kind, changedBy)
		})
	default:
		return fmt.Errorf("unsupported file type %s", kind)
	}
}

func ensureOrderID(orderID int, ids []int) error {
	for _, id := range ids {
		if id != orderID {
			return fmt.Errorf("file contains order_id %d, expected %d", id, orderID)
		}
	}
	return nil
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
	report, err := requestAllOrdersReportFromOneC()
	if err != nil {
		writeOneCError(w, err)
		return
	}

	file, err := generateExcel(report.Date)
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

func requestAllOrdersReportFromOneC() (models.ReportRequest, error) {
	if strings.TrimSpace(handlerConfig.OneCURL) == "" {
		return buildLocalOrdersReport()
	}

	return requestEmptyReportFromOneC()
}

func buildLocalOrdersReport() (models.ReportRequest, error) {
	orders, err := storage.GetOrders()
	if err != nil {
		return models.ReportRequest{}, err
	}

	items := make([]models.ReportItem, 0, len(orders)*4)
	for _, order := range orders {
		cost, err := storage.CalculateCost(order.Id, "bom")
		if err != nil {
			return models.ReportRequest{}, err
		}
		orderID := strconv.Itoa(order.Id)
		items = append(items,
			models.ReportItem{Order: orderID, Type: "Материалы", Sum: cost.Materials},
			models.ReportItem{Order: orderID, Type: "Труд", Sum: cost.Labor},
			models.ReportItem{Order: orderID, Type: "Накладные", Sum: cost.Overhead},
			models.ReportItem{Order: orderID, Type: "Себестоимость", Sum: cost.Total},
		)
	}
	return models.ReportRequest{
		Status: "local",
		Date:   items,
	}, nil
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
