package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func GenerateReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var generateReq models.GenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&generateReq); err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusBadRequest)
		return
	}
	if generateReq.CalculationID == 0 {
		generateReq.CalculationID = generateReq.CalculationIDCamel
	}
	if generateReq.CalculationID == 0 {
		http.Error(w, "Не указан id расчета", http.StatusBadRequest)
		return
	}

	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Не удалось определить пользователя", http.StatusUnauthorized)
		return
	}

	reqData, err := storage.GetReportCalculation(generateReq.CalculationID, userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Расчет не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Не удалось получить расчет", http.StatusInternalServerError)
		return
	}

	url := os.Getenv("ONEC_URL")
	if url == "" {
		http.Error(w, "Не задан URL 1С", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		http.Error(w, "Не удалось обработать данные", http.StatusBadRequest)
		return
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		http.Error(w, "Ошибка создания запроса к 1С", http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	status, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, "Ошибка соединения с 1С", http.StatusBadGateway)
		return
	}
	defer status.Body.Close()

	if status.StatusCode != http.StatusOK {
		http.Error(w, "Ошибка от 1С", http.StatusBadGateway)
		return
	}

	var reportData models.ReportRequest
	if err = json.NewDecoder(status.Body).Decode(&reportData); err != nil {
		http.Error(w, "Не удалось получить данные из 1С", http.StatusBadRequest)
		return
	}

	reportID, err := storage.CreateReport(generateReq.CalculationID, userID, reportData)
	if err != nil {
		http.Error(w, "Не удалось сохранить отчет", http.StatusInternalServerError)
		return
	}

	file, err := generateExcel(reportData.Date)
	if err != nil {
		http.Error(w, "Ошибка при создании файла Excel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
	w.Header().Set("X-Report-ID", fmt.Sprintf("%d", reportID))

	if err := file.Write(w); err != nil {
		http.Error(w, "Ошибка при создании Excel", http.StatusInternalServerError)
	}
}

func DownloadReportExcelHandler(w http.ResponseWriter, r *http.Request) {
	reportID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Неверный id отчета", http.StatusBadRequest)
		return
	}

	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Не удалось определить пользователя", http.StatusUnauthorized)
		return
	}

	report, err := storage.GetReportByID(reportID, userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Отчет не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Не удалось получить отчет", http.StatusInternalServerError)
		return
	}

	file, err := generateExcel(report.OneCReportData.Date)
	if err != nil {
		http.Error(w, "Ошибка при создании файла Excel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")

	if err := file.Write(w); err != nil {
		http.Error(w, "Ошибка при создании Excel", http.StatusInternalServerError)
	}
}

func generateExcel(items []models.ReportItem) (*excelize.File, error) {
	f := excelize.NewFile()
	sheetName := f.GetSheetName(0)
	f.SetCellValue(sheetName, "A1", "Order")
	f.SetCellValue(sheetName, "B1", "Type")
	f.SetCellValue(sheetName, "C1", "Sum")

	for i, item := range items {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item.Order)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Type)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Sum)
	}
	return f, nil
}
