package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

func GenerateReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var reqData struct{ OrderID int }
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", status.Header.Get("Content-Type"))
	w.Header().Set("Content-Disposition", status.Header.Get("Content-Disposition"))

	io.Copy(w, status.Body)
}
