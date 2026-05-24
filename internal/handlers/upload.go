package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/parser"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/utils"
)

func DoTransactions(insertFunc func(*sql.Tx) error) error {
	tx, err := storage.DB().Begin()
	if err != nil {
		return fmt.Errorf("Ошибка начала транзакции %w", err)
	}

	err = insertFunc(tx)
	if err != nil {
		if csvErr, ok := err.(*models.CSVError); ok {
			tx.Rollback()
			storage.SaveError(csvErr)
			return fmt.Errorf("Ошибка вставки данных %w", err)
		}
		tx.Rollback()
		return fmt.Errorf("Ошибка вставки данных %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Ошибка завершения транзакции %w", err)
	}
	return nil
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(handlerConfig.UploadMaxMemory)
	if err != nil {
		http.Error(w, "Ошибка парсинга формы "+err.Error(), http.StatusInternalServerError)
		return
	}

	typeFile := r.FormValue("type")
	changedBy := r.Header.Get("X-User")
	if changedBy == "" {
		changedBy = "unknown"
	}
	file, _, err := r.FormFile("file")

	if err != nil {
		http.Error(w, "Ошибка получения файла "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	temp, err := os.CreateTemp("", "*")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла "+err.Error(), http.StatusInternalServerError)
		return
	}

	defer temp.Close()
	defer os.Remove(temp.Name())

	_, err = io.Copy(temp, file)
	if err != nil {
		http.Error(w, "Ошибка копирования данных "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch typeFile {
	case "boms":
		items, err := parser.ParseBOM(temp.Name())
		if err != nil {
			if csvErr, ok := err.(*models.CSVError); ok {
				storage.SaveError(csvErr)
				http.Error(w, "Ошибка парсинга BOM "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
			return
		}
		orderIDs := utils.ExtractOrderIds(items)
		err = storage.ValidateOrders(orderIDs)
		if err != nil {
			storage.SaveError(&models.CSVError{
				FileName: typeFile,
				Row:      0,
				Column:   "",
				Cause:    err.Error(),
			})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = DoTransactions(func(tx *sql.Tx) error {
			if err := storage.InsertBOMItems(tx, items); err != nil {
				return err
			}
			return storage.SaveUploadLog(tx, orderIDs, typeFile, changedBy)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "labor":
		items, err := parser.ParseLabor(temp.Name())
		if err != nil {
			if csvErr, ok := err.(*models.CSVError); ok {
				storage.SaveError(csvErr)
				http.Error(w, "Ошибка парсинга Labor "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
			return
		}
		orderIDs := utils.ExtractOrderIds(items)
		err = storage.ValidateOrders(orderIDs)
		if err != nil {
			storage.SaveError(&models.CSVError{
				FileName: typeFile,
				Row:      0,
				Column:   "",
				Cause:    err.Error(),
			})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = DoTransactions(func(tx *sql.Tx) error {
			if err := storage.InsertLaborItems(tx, items); err != nil {
				return err
			}
			return storage.SaveUploadLog(tx, orderIDs, typeFile, changedBy)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case "overhead":
		items, err := parser.ParseOverhead(temp.Name())
		if err != nil {
			if csvErr, ok := err.(*models.CSVError); ok {
				storage.SaveError(csvErr)
				http.Error(w, "Ошибка парсинга Overhead "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = DoTransactions(func(tx *sql.Tx) error {
			return storage.InsertOverheadItems(tx, items)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Некорректный тип файла", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
}
