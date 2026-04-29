package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/config"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/storage"

	"github.com/golang-jwt/jwt/v5"
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
	if strings.Count(req.Email, "@") != 1 {
		http.Error(w, "Неверный формат почты", http.StatusBadRequest)
		return
	}

	hash, err := storage.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Не удалось получить пароль", http.StatusBadRequest)
		return
	}

	var exists int
	err = storage.DB().QueryRow("SELECT 1 FROM users WHERE email=$1", req.Email).Scan(&exists)
	if err == nil {
		http.Error(w, "Пользователь с таким email уже зарегестрирован", http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		http.Error(w, "Не удалось получить данные из базы данных", http.StatusInternalServerError)
		return
	}
	id, err := storage.CreateUser(req.Email, hash, "user")
	if err != nil {
		http.Error(w, "Не удалось создать пользователя", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok", "id": id})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Не удалось получить данные", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Не указан пароль или логин", http.StatusBadRequest)
		return
	}
	if strings.Count(req.Email, "@") != 1 {
		http.Error(w, "Неверный формат почты", http.StatusBadRequest)
		return
	}

	user, err := storage.GetUserEmail(req.Email)
	if err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}
	if !storage.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Не удалось создать jwt-токен", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"status": "ok", "token": token})
}

func generateToken(userID int, email string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.GetJWTSecret())
}
