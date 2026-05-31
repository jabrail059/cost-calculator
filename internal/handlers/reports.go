package handlers

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

	userID, hasUser := optionalUserIDFromRequest(r)

	reqData, err := reportCalculationForRequest(generateReq.CalculationID, userID, hasUser)
	if err == sql.ErrNoRows {
		http.Error(w, "Расчет не найден", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Не удалось получить расчет", http.StatusInternalServerError)
		return
	}

	reportData, err := requestReportFromOneC(r, reqData)
	if err != nil {
		writeOneCError(w, err)
		return
	}

	reportOwnerID := userID
	if !hasUser {
		reportOwnerID, err = storage.GetOrCreateAnonymousUserID()
		if err != nil {
			http.Error(w, "Не удалось подготовить пользователя для сохранения отчета", http.StatusInternalServerError)
			return
		}
	}

	reportID, err := storage.CreateReport(generateReq.CalculationID, reportOwnerID, reportData)
	if err != nil {
		http.Error(w, "Не удалось сохранить отчет", http.StatusInternalServerError)
		return
	}

	file, err := generateExcel(reportData.Date)
	if err != nil {
		http.Error(w, "Ошибка при создании файла Excel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
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

	userID, hasUser := optionalUserIDFromRequest(r)

	var report *models.Report
	if hasUser {
		report, err = storage.GetReportByID(reportID, userID)
	} else {
		report, err = storage.GetReportByIDPublic(reportID)
	}

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

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")

	if err := file.Write(w); err != nil {
		http.Error(w, "Ошибка при создании Excel", http.StatusInternalServerError)
	}
}

func GenerateExcelFromOneCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var reportData models.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&reportData); err != nil {
		http.Error(w, "Не удалось прочитать JSON отчета", http.StatusBadRequest)
		return
	}
	if len(reportData.Date) == 0 {
		http.Error(w, "Пустой отчет", http.StatusBadRequest)
		return
	}

	file, err := generateExcel(reportData.Date)
	if err != nil {
		http.Error(w, "Ошибка при создании файла Excel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
	if err := file.Write(w); err != nil {
		http.Error(w, "Ошибка при создании Excel", http.StatusInternalServerError)
	}
}

func reportCalculationForRequest(calculationID int, userID int, hasUser bool) (*models.CalculationResult, error) {
	if hasUser {
		data, err := storage.GetReportCalculation(calculationID, userID)
		if err == nil {
			return data, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	return storage.GetReportCalculationByID(calculationID)
}

func requestReportFromOneC(r *http.Request, reqData *models.CalculationResult) (models.ReportRequest, error) {
	if strings.TrimSpace(handlerConfig.OneCURL) == "" {
		return buildLocalReport(reqData), nil
	}

	return requestEmptyReportFromOneC()
}

func requestJSONReportFromOneC(payload any) (models.ReportRequest, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return models.ReportRequest{}, fmt.Errorf("Не удалось подготовить данные для 1С")
	}
	requestBody := string(jsonData)

	httpReq, err := http.NewRequest(http.MethodPost, handlerConfig.OneCURL, bytes.NewReader(jsonData))
	if err != nil {
		return models.ReportRequest{}, fmt.Errorf("Ошибка создания запроса к 1С")
	}
	httpReq.ContentLength = int64(len(jsonData))
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: handlerConfig.OneCTimeout,
		Transport: &http.Transport{
			Proxy: nil,
			TLSClientConfig: &tls.Config{
				Renegotiation: tls.RenegotiateOnceAsClient,
			},
		},
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return models.ReportRequest{}, &oneCError{
			Message:     "Ошибка соединения с 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: requestBody,
			Cause:       err.Error(),
		}
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return models.ReportRequest{}, &oneCError{
			Message:     "Не удалось прочитать ответ 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: requestBody,
			StatusCode:  resp.StatusCode,
			Status:      resp.Status,
			Headers:     resp.Header,
			Cause:       readErr.Error(),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return models.ReportRequest{}, &oneCError{
			Message:     "Ошибка от 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: requestBody,
			StatusCode:  resp.StatusCode,
			Status:      resp.Status,
			Headers:     resp.Header,
			Body:        trimLongResponse(body),
		}
	}

	var reportData models.ReportRequest
	if err := json.Unmarshal(body, &reportData); err != nil {
		return models.ReportRequest{}, &oneCError{
			Message:     "Не удалось разобрать JSON ответа 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: requestBody,
			StatusCode:  resp.StatusCode,
			Status:      resp.Status,
			Headers:     resp.Header,
			Body:        trimLongResponse(body),
			Cause:       err.Error(),
		}
	}

	if len(reportData.Date) == 0 {
		return models.ReportRequest{}, fmt.Errorf("1С вернула пустой отчет")
	}

	return reportData, nil
}

func requestEmptyReportFromOneC() (models.ReportRequest, error) {
	httpReq, err := http.NewRequest(http.MethodPost, handlerConfig.OneCURL, http.NoBody)
	if err != nil {
		return models.ReportRequest{}, fmt.Errorf("Ошибка создания запроса к 1С")
	}
	httpReq.ContentLength = 0
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Length", "0")

	client := &http.Client{
		Timeout: handlerConfig.OneCTimeout,
		Transport: &http.Transport{
			Proxy: nil,
			TLSClientConfig: &tls.Config{
				Renegotiation: tls.RenegotiateOnceAsClient,
			},
		},
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return models.ReportRequest{}, &oneCError{
			Message:     "Ошибка соединения с 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: "",
			Cause:       err.Error(),
		}
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return models.ReportRequest{}, &oneCError{
			Message:     "Не удалось прочитать ответ 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: "",
			StatusCode:  resp.StatusCode,
			Status:      resp.Status,
			Headers:     resp.Header,
			Cause:       readErr.Error(),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return models.ReportRequest{}, &oneCError{
			Message:     "Ошибка от 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: "",
			StatusCode:  resp.StatusCode,
			Status:      resp.Status,
			Headers:     resp.Header,
			Body:        trimLongResponse(body),
		}
	}

	var reportData models.ReportRequest
	if err := json.Unmarshal(body, &reportData); err != nil {
		return models.ReportRequest{}, &oneCError{
			Message:     "Не удалось разобрать JSON ответа 1С",
			RequestURL:  handlerConfig.OneCURL,
			RequestBody: "",
			StatusCode:  resp.StatusCode,
			Status:      resp.Status,
			Headers:     resp.Header,
			Body:        trimLongResponse(body),
			Cause:       err.Error(),
		}
	}

	if len(reportData.Date) == 0 {
		return models.ReportRequest{}, fmt.Errorf("1С вернула пустой отчет")
	}

	return reportData, nil
}

type oneCError struct {
	Message     string      `json:"message"`
	RequestURL  string      `json:"request_url,omitempty"`
	RequestBody string      `json:"request_body,omitempty"`
	StatusCode  int         `json:"onec_status_code,omitempty"`
	Status      string      `json:"onec_status,omitempty"`
	Headers     http.Header `json:"onec_headers,omitempty"`
	Body        string      `json:"onec_body,omitempty"`
	Cause       string      `json:"cause,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

func (e *oneCError) Error() string {
	if e.Cause != "" {
		return e.Message + ": " + e.Cause
	}
	if e.Status != "" {
		return e.Message + ": " + e.Status
	}
	return e.Message
}

func writeOneCError(w http.ResponseWriter, err error) {
	diag, ok := err.(*oneCError)
	if !ok {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	diag.Timestamp = time.Now().UTC()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(diag)
}

func buildLocalReport(calc *models.CalculationResult) models.ReportRequest {
	order := fmt.Sprintf("Расчет %d", calc.CalculationID)
	return models.ReportRequest{
		Status: "local",
		Date: []models.ReportItem{
			{Order: order, Type: "BOM", Sum: calc.BomCost},
			{Order: order, Type: "Labor", Sum: calc.LaborCost},
			{Order: order, Type: "Overhead", Sum: calc.OverheadCost},
			{Order: order, Type: "Total", Sum: calc.TotalCost},
		},
	}
}

func trimLongResponse(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > 1000 {
		return text[:1000] + "..."
	}
	return text
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
