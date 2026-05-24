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
		writeJSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Не удалось получить данные", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "Не указан логин или пароль", http.StatusBadRequest)
		return
	}
	if strings.Count(req.Email, "@") != 1 {
		writeJSONError(w, "Неверный формат почты", http.StatusBadRequest)
		return
	}

	hash, err := storage.HashPassword(req.Password)
	if err != nil {
		writeJSONError(w, "Не удалось получить пароль", http.StatusBadRequest)
		return
	}

	var exists int
	err = storage.DB().QueryRow("SELECT 1 FROM users WHERE email=$1", req.Email).Scan(&exists)
	if err == nil {
		writeJSONError(w, "Пользователь с таким email уже зарегистрирован", http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		writeJSONError(w, "Не удалось получить данные из базы данных", http.StatusInternalServerError)
		return
	}

	id, err := storage.CreateUser(req.Email, hash, "user")
	if err != nil {
		writeJSONError(w, "Не удалось создать пользователя", http.StatusInternalServerError)
		return
	}

	token, err := generateToken(id, req.Email, "user")
	if err != nil {
		writeJSONError(w, "Не удалось создать jwt-токен", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"status": "ok", "id": id, "token": token})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Не удалось получить данные", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "Не указан пароль или логин", http.StatusBadRequest)
		return
	}
	if strings.Count(req.Email, "@") != 1 {
		writeJSONError(w, "Неверный формат почты", http.StatusBadRequest)
		return
	}

	user, err := storage.GetUserEmail(req.Email)
	if err != nil {
		writeJSONError(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}
	if !storage.CheckPasswordHash(req.Password, user.PasswordHash) {
		writeJSONError(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID, user.Email, user.Role)
	if err != nil {
		writeJSONError(w, "Не удалось создать jwt-токен", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "token": token})
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
