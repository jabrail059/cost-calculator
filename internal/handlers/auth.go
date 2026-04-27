package handlers

import (
	"encoding/json"
	"net/http"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	//"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Не указан логин или пароль", http.StatusBadRequest)
		return
	}

	// hash, err := storage.HashPassword(req.Password)
	// if err != nil {
	// 	http.Error(w, "", http.StatusBadRequest)
	// }

}
