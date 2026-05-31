package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/config"
)

func userIDFromRequest(r *http.Request) (int, error) {
	claims, ok := r.Context().Value("user").(jwt.MapClaims)
	if ok {
		return userIDFromClaims(claims)
	}

	claims, err := claimsFromAuthorizationHeader(r)
	if err != nil {
		return 0, err
	}
	return userIDFromClaims(claims)
}

func optionalUserIDFromRequest(r *http.Request) (int, bool) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func requestWithUserClaims(r *http.Request) *http.Request {
	if _, ok := r.Context().Value("user").(jwt.MapClaims); ok {
		return r
	}

	claims, err := claimsFromAuthorizationHeader(r)
	if err != nil {
		return r
	}

	ctx := context.WithValue(r.Context(), "user", claims)
	return r.WithContext(ctx)
}

func claimsFromAuthorizationHeader(r *http.Request) (jwt.MapClaims, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header not found")
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenString == "" || tokenString == authHeader && strings.HasPrefix(strings.ToLower(authHeader), "bearer") {
		return nil, fmt.Errorf("token not found")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.GetJWTSecret(), nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

func userIDFromClaims(claims jwt.MapClaims) (int, error) {
	switch id := claims["user_id"].(type) {
	case float64:
		return int(id), nil
	case int:
		return id, nil
	case jsonNumber:
		parsed, err := id.Int64()
		if err != nil {
			return 0, fmt.Errorf("invalid user id")
		}
		return int(parsed), nil
	default:
		return 0, fmt.Errorf("invalid user id")
	}
}

type jsonNumber interface {
	Int64() (int64, error)
}
