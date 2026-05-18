package handlers

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func userIDFromRequest(r *http.Request) (int, error) {
	claims, ok := r.Context().Value("user").(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("user claims not found")
	}

	switch id := claims["user_id"].(type) {
	case float64:
		return int(id), nil
	case int:
		return id, nil
	default:
		return 0, fmt.Errorf("invalid user id")
	}
}
