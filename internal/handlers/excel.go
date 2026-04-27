package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/xuri/excelize/v2"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

func GenerateExcelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	var req models.ReportRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Не удалось получить данные из 1С", http.StatusBadRequest)
		return
	}

	file, err := generateExcel(req.Date)
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
