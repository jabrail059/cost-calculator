package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func DoTransaktions(insertFunc func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Ошибка начала транзакции %w", err)
	}

	err = insertFunc(tx)
	if err != nil {
		if csvErr, ok := err.(*CSVError); ok {
			tx.Rollback()
			SaveError(csvErr)
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
	if r.Method == "POST" {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Ошибка парсинга формы "+err.Error(), http.StatusInternalServerError)
			return
		}

		typeFile := r.FormValue("type")
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
			items, err := ParseBOM(temp.Name())
			if err != nil {
				if csvErr, ok := err.(*CSVError); ok {
					SaveError(csvErr)
					http.Error(w, "Ошибка парсинга BOM "+err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
				return
			}
			err = ValidateOrders(ExtractOrderIds(items))
			if err != nil {
				SaveError(&CSVError{
					FileName: typeFile,
					Row:      0,
					Column:   "",
					Cause:    err.Error(),
				})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = DoTransaktions(func(tx *sql.Tx) error {
				return InsertBOMItems(tx, items)
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case "labor":
			items, err := ParseLabor(temp.Name())
			if err != nil {
				if csvErr, ok := err.(*CSVError); ok {
					SaveError(csvErr)
					http.Error(w, "Ошибка парсинга Labor "+err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
				return
			}
			err = ValidateOrders(ExtractOrderIds(items))
			if err != nil {
				SaveError(&CSVError{
					FileName: typeFile,
					Row:      0,
					Column:   "",
					Cause:    err.Error(),
				})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = DoTransaktions(func(tx *sql.Tx) error {
				return InsertLaborItems(tx, items)
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case "overhead":
			items, err := ParseOverhead(temp.Name())
			if err != nil {
				if csvErr, ok := err.(*CSVError); ok {
					SaveError(csvErr)
					http.Error(w, "Ошибка парсинга Overhead "+err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, "Ошибка "+err.Error(), http.StatusInternalServerError)
				return
			}
			err = ValidateOrders(ExtractOrderIds(items))
			if err != nil {
				SaveError(&CSVError{
					FileName: typeFile,
					Row:      0,
					Column:   "",
					Cause:    err.Error(),
				})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = DoTransaktions(func(tx *sql.Tx) error {
				return InsertOverheadItems(tx, items)
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
	} else {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
}

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select id, start_date, end_date, total_cost, status, error_id from orders")
	if err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	orders := []OrderResponse{}
	for rows.Next() {
		order := OrderResponse{}
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
	rows := db.QueryRow("select id, start_date, end_date, total_cost, status, error_id from orders where id = $1", id)
	order := OrderResponse{}
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
	costResp, err := CalculateCost(id)
	if err != nil {
		http.Error(w, "Не удалось посчитать себестоимость", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costResp)
}
