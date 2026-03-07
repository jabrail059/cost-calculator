package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func DoTransaktions(insertFunc func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Ошибка начала транзакции %w", err)
	}

	err = insertFunc(tx)
	if err != nil {
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
			http.Error(w, "Ошибка парсинга формы "+err.Error(), http.StatusBadRequest)
			return
		}

		typeFile := r.FormValue("type")
		file, _, err := r.FormFile("file")

		if err != nil {
			http.Error(w, "Ошибка получения файла "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		temp, err := os.CreateTemp("", "*")
		if err != nil {
			http.Error(w, "Ошибка создания временного файла "+err.Error(), http.StatusBadRequest)
			return
		}

		defer temp.Close()
		defer os.Remove(temp.Name())

		_, err = io.Copy(temp, file)
		if err != nil {
			http.Error(w, "Ошибка копирования данных "+err.Error(), http.StatusBadRequest)
			return
		}

		switch typeFile {
		case "boms":
			items, err := ParseBOM(temp.Name())
			if err != nil {
				http.Error(w, "Ошибка парсинга BOM "+err.Error(), http.StatusBadRequest)
				return
			}
			err = DoTransaktions(func(tx *sql.Tx) error {
				return InsertBOMItems(tx, items)
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		case "labor":
			items, err := ParseLabor(temp.Name())
			if err != nil {
				http.Error(w, "Ошибка парсинга Labor "+err.Error(), http.StatusBadRequest)
				return
			}
			err = DoTransaktions(func(tx *sql.Tx) error {
				return InsertLaborItems(tx, items)
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		case "overhead":
			items, err := ParseOverhead(temp.Name())
			if err != nil {
				http.Error(w, "Ошибка парсинга Overhead "+err.Error(), http.StatusBadRequest)
				return
			}
			err = DoTransaktions(func(tx *sql.Tx) error {
				return InsertOverheadItems(tx, items)
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
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
