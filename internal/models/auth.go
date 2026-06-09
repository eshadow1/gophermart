// Package models определяет основные модели данных, бизнес-ошибки и ключи контекста,
// используемые во всех слоях приложения (handler, service, repository).
package models

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidFormat обозначает невалидный формат запроса
	ErrInvalidFormat = errors.New("invalid request format")
	// ErrUserAlreadyExists обозначает конфликт: пользователь с таким логином
	// уже зарегистрирован в системе.
	ErrUserAlreadyExists = errors.New("login already exists")
	// ErrInvalidCredentials обозначает ошибку аутентификации:
	// неверный логин или пароль.
	ErrInvalidCredentials = errors.New("invalid login or password")
	// ErrUnauthorized обозначает отсутствие авторизации:
	// токен отсутствует, невалиден или просрочен.
	ErrUnauthorized = errors.New("unauthorized")
)

type contextKey string

const (
	// UserIDContextKey — ключ для сохранения и извлечения идентификатора
	// пользователя из context.Context.
	UserIDContextKey contextKey = "user_id"
)

// UserClaims описывает полезную нагрузку (claims) JWT-токена пользователя.
type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// User представляет модель пользователя для работы с базой данных.
type User struct {
	// ID — уникальный идентификатор пользователя.
	ID int64
	// Login — логин пользователя (должен быть уникальным).
	Login string
	// PasswordHash — хэш пароля пользователя.
	PasswordHash string
}

// AuthCredentials представляет модель для декодирования JSON-тела
// из HTTP-запросов регистрации и входа пользователя.
type AuthCredentials struct {
	// Login — логин пользователя.
	Login string `json:"login"`
	// Password — пароль пользователя в открытом виде.
	Password string `json:"password"`
}
