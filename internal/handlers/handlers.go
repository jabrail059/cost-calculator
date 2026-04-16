package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
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
		if csvErr, ok := err.(*storage.CSVError); ok {
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
	err := r.ParseMultipartForm(10 << 20)
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
			if csvErr, ok := err.(*storage.CSVError); ok {
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
			storage.SaveError(&storage.CSVError{
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
			if csvErr, ok := err.(*storage.CSVError); ok {
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
			storage.SaveError(&storage.CSVError{
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
			if csvErr, ok := err.(*storage.CSVError); ok {
				storage.SaveError(csvErr)
				http.Error(w, "Ошибка парсинга Overhead "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
			return
		}
		orderIDs := utils.ExtractOrderIds(items)
		err = storage.ValidateOrders(orderIDs)
		if err != nil {
			storage.SaveError(&storage.CSVError{
				FileName: typeFile,
				Row:      0,
				Column:   "",
				Cause:    err.Error(),
			})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = DoTransactions(func(tx *sql.Tx) error {
			if err := storage.InsertOverheadItems(tx, items); err != nil {
				return err
			}
			return storage.SaveUploadLog(tx, orderIDs, typeFile, changedBy)
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

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := storage.DB().Query("select id, start_date, end_date, total_cost, status, error_id from orders")
	if err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	orders := []models.OrderResponse{}
	for rows.Next() {
		order := models.OrderResponse{}
		err = rows.Scan(&order.Id, &order.StartDate, &order.EndDate, &order.TotalCost, &order.Status, &order.ErrorId)
		if err != nil {
			http.Error(w, "Не удалось считать данные", http.StatusInternalServerError)
			return
		}
		orders = append(orders, order)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func GetOrderByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	rows := storage.DB().QueryRow("select id, start_date, end_date, total_cost, status, error_id from orders where id = $1", id)
	order := models.OrderResponse{}
	err := rows.Scan(&order.Id, &order.StartDate, &order.EndDate, &order.TotalCost, &order.Status, &order.ErrorId)
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
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный формат идентификатора заказа", http.StatusBadRequest)
		return
	}
	costResp, err := storage.CalculateCost(id)
	if err != nil {
		http.Error(w, "Не удалось посчитать себестоимость", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costResp)
}

func CalculateFromFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(10 << 20)
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
		if csvErr, ok := err.(*storage.CSVError); ok {
			storage.SaveError(csvErr)
			http.Error(w, "Ошибка парсинга BOM "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
		return
	}

	laborItems, err := parser.ParseLabor(laborTemp.Name())
	if err != nil {
		if csvErr, ok := err.(*storage.CSVError); ok {
			storage.SaveError(csvErr)
			http.Error(w, "Ошибка парсинга Labor "+err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
		return
	}

	overheadItems, err := parser.ParseOverhead(overheadTemp.Name())
	if err != nil {
		if csvErr, ok := err.(*storage.CSVError); ok {
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func GetOrderBOMsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный идентификатор заказа", http.StatusBadRequest)
		return
	}

	rows, err := storage.DB().Query("SELECT id, order_id, quantity, unit_cost, material_code FROM boms WHERE order_id = $1", id)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.BOMItemResponse
	for rows.Next() {
		var item models.BOMItemResponse
		err := rows.Scan(&item.ID, &item.OrderID, &item.Quantity, &item.UnitCost, &item.MaterialCode)
		if err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func GetOrderLaborHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный идентификатор заказа", http.StatusBadRequest)
		return
	}

	rows, err := storage.DB().Query("SELECT id, order_id, rate, hours FROM labor WHERE order_id = $1", id)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.LaborItemResponse
	for rows.Next() {
		var item models.LaborItemResponse
		err := rows.Scan(&item.ID, &item.OrderID, &item.Rate, &item.Hours)
		if err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func GetOrderOverheadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный идентификатор заказа", http.StatusBadRequest)
		return
	}

	rows, err := storage.DB().Query("SELECT id, order_id, date, prod_type, amount FROM overhead WHERE order_id = $1", id)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.OverheadItemResponse
	for rows.Next() {
		var item models.OverheadItemResponse
		err := rows.Scan(&item.ID, &item.OrderID, &item.Date, &item.ProdType, &item.Amount)
		if err != nil {
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
