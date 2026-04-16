package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func GetOrderChangesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	IdStr := vars["id"]
	Id, err := strconv.Atoi(IdStr)
	if err != nil {
		http.Error(w, "Не удалось получить ", http.StatusBadRequest)
		return
	}
	rows, err := storage.DB().Query("select id, order_id, file_type, uploaded_at, changed_by FROM upload_log WHERE order_id = $1 ORDER BY uploaded_at DESC", Id)
	if err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []models.Log
	for rows.Next() {
		log := models.Log{}
		err := rows.Scan(&log.Id, &log.OrderID, &log.Filetype, &log.UploadedAt, &log.ChangedBy)
		if err != nil {
			http.Error(w, "Не удалось считать данные", http.StatusInternalServerError)
			return
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Возникла ошибка при считывании данных", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
