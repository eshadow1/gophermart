// Package middleware предоставляет HTTP-middleware для проверки аутентификации
// пользователей с помощью JWT-токенов. Middleware извлекает токен из заголовка
// Authorization или HTTP-куки, валидирует его и добавляет идентификатор
// пользователя в контекст запроса.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware создает middleware для проверки JWT-токена и авторизации пользователя.
func AuthMiddleware(cfg *configs.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tokenStr := extractToken(r)
			if tokenStr == "" {
				utils.RespondError(w, &utils.UnauthorizedError{})
				return
			}

			userID, err := parseToken(tokenStr, cfg.JWTSecret)
			if err != nil {
				utils.RespondError(w, &utils.UnauthorizedError{})
				return
			}

			ctx = context.WithValue(ctx, models.UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken извлекает строку JWT-токена из HTTP-запроса.
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	c, err := r.Cookie("auth_token")
	if err == nil {
		return c.Value
	}
	return ""
}

// parseToken парсит и валидирует JWT-токен, извлекая из него идентификатор пользователя.
func parseToken(tokenStr string, jwtSecret []byte) (int64, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(*jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, models.ErrUnauthorized
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, models.ErrUnauthorized
	}
	return int64(sub), nil
}
