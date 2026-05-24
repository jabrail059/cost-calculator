package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/parser"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func CalculateFromFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(handlerConfig.UploadMaxMemory)
	if err != nil {
		http.Error(w, "Ошибка парсинга формы "+err.Error(), http.StatusInternalServerError)
		return
	}
	bomFile, _, err := r.FormFile("bom")
	if err != nil {
		http.Error(w, "Не удалось загрузить файл Bom "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer bomFile.Close()

	laborFile, _, err := r.FormFile("labor")
	if err != nil {
		http.Error(w, "Не удалось загрузить файл Labor "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer laborFile.Close()

	overheadFile, _, err := r.FormFile("overhead")
	if err != nil {
		http.Error(w, "Не удалось загрузить файл Overhead "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer overheadFile.Close()

	bomTemp, err := os.CreateTemp("", "*")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer bomTemp.Close()
	defer os.Remove(bomTemp.Name())

	_, err = io.Copy(bomTemp, bomFile)
	if err != nil {
		http.Error(w, "Ошибка копирования данных "+err.Error(), http.StatusInternalServerError)
		return
	}
	laborTemp, err := os.CreateTemp("", "*")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer laborTemp.Close()
	defer os.Remove(laborTemp.Name())

	_, err = io.Copy(laborTemp, laborFile)
	if err != nil {
		http.Error(w, "Ошибка копирования данных "+err.Error(), http.StatusInternalServerError)
		return
	}
	overheadTemp, err := os.CreateTemp("", "*")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer overheadTemp.Close()
	defer os.Remove(overheadTemp.Name())

	_, err = io.Copy(overheadTemp, overheadFile)
	if err != nil {
		http.Error(w, "Ошибка копирования данных "+err.Error(), http.StatusInternalServerError)
		return
	}

	bomItems, err := parser.ParseBOM(bomTemp.Name())
	if err != nil {
		if csvErr, ok := err.(*models.CSVError); ok {
			storage.SaveError(csvErr)
			http.Error(w, "Ошибка парсинга BOM "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
		return
	}

	laborItems, err := parser.ParseLabor(laborTemp.Name())
	if err != nil {
		if csvErr, ok := err.(*models.CSVError); ok {
			storage.SaveError(csvErr)
			http.Error(w, "Ошибка парсинга Labor "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
		return
	}

	overheadItems, err := parser.ParseOverhead(overheadTemp.Name())
	if err != nil {
		if csvErr, ok := err.(*models.CSVError); ok {
			storage.SaveError(csvErr)
			http.Error(w, "Ошибка парсинга Overhead "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
		return
	}

	var bomTotal, laborTotal, overheadTotal float64
	for _, item := range bomItems {
		bomTotal += item.Quantity * item.UnitCost
	}

	for _, item := range laborItems {
		laborTotal += item.Hours * item.Rate
	}

	for _, item := range overheadItems {
		overheadTotal += item.Amount
	}

	result := models.CalculationResult{
		BomCost:      bomTotal,
		LaborCost:    laborTotal,
		OverheadCost: overheadTotal,
		TotalCost:    bomTotal + laborTotal + overheadTotal,
	}

	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Не удалось определить пользователя", http.StatusUnauthorized)
		return
	}

	calculationID, err := storage.CreateReportCalculation(userID, result)
	if err != nil {
		http.Error(w, "Не удалось сохранить расчет", http.StatusInternalServerError)
		return
	}
	result.CalculationID = calculationID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
