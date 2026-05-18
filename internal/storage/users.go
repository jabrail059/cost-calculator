package storage

import (
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(email string, passwordHash string, role string) (int, error) {
	var id int
	err := db.QueryRow("INSERT INTO users(email, password_hash, role) VALUES ($1, $2, $3) RETURNING id",
		email,
		passwordHash,
		role).Scan(&id)
	return id, err
}

func GetUserEmail(email string) (*models.User, error) {
	var user models.User
	err := db.QueryRow("SELECT id, email, password_hash, role, created_at FROM users WHERE email=$1",
		email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetUsersByEmails(emails []string) ([]models.User, error) {
	var users []models.User
	for _, email := range emails {
		var user models.User
		if err := db.QueryRow("SELECT id, email, password_hash, role, created_at FROM users WHERE email=$1",
			email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
