// Package handler предоставляет HTTP-хендлеры для обработки запросов аутентификации,
// таких как регистрация и вход пользователей. Пакет инкапсулирует работу с HTTP-протоколом
// и делегирует бизнес-логику интерфейсу AuthService.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/utils"
)

// AuthService определяет контракт для слоя бизнес-логики аутентификации.
// Реализации этого интерфейса должны обрабатывать регистрацию и вход пользователей,
// возвращая JWT-токен в случае успеха или специфичную ошибку бизнес-логики.
type AuthService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

// authHandler реализует HTTP-хендлеры для эндпоинтов аутентификации.
type authHandler struct {
	svc AuthService
}

// NewAuthHandler создает и возвращает новый экземпляр HTTP-хендлера для аутентификации.
// В качестве зависимости принимает реализацию интерфейса AuthService.
func NewAuthHandler(svc AuthService) *authHandler {
	return &authHandler{svc: svc}
}

func (*authHandler) respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidFormat):
		utils.RespondError(w, &utils.ValidationError{Msg: err.Error()})
	case errors.Is(err, models.ErrUserAlreadyExists):
		utils.RespondError(w, &utils.ConflictError{Msg: err.Error()})
	case errors.Is(err, models.ErrInvalidCredentials):
		utils.RespondError(w, &utils.UnauthorizedError{Msg: err.Error()})
	default:
		utils.RespondError(w, err)
	}
}

// RegisterUser обрабатывает HTTP-запросы на регистрацию нового пользователя.
// Ожидает, что тело запроса содержит JSON с полями login и password.
func (h *authHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	h.handleAuthAction(w, r, h.svc.Register)
}

// LoginUser обрабатывает HTTP-запросы на вход пользователя в систему.
// Ожидает, что тело запроса содержит JSON с полями login и password.
// В случае успеха устанавливает HTTP-куку auth_token с JWT-токеном и возвращает статус 200 OK.
func (h *authHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	h.handleAuthAction(w, r, h.svc.Login)
}

// authFunc описывает сигнатуру методов Register и Login вашего сервиса
type authFunc func(ctx context.Context, login, password string) (string, error)

func (h *authHandler) handleAuthAction(w http.ResponseWriter, r *http.Request, action authFunc) {
	var creds models.AuthCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		h.respondError(w, models.ErrInvalidFormat)
		return
	}

	token, err := action(r.Context(), strings.TrimSpace(creds.Login), creds.Password)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.setAuthCookie(w, token)
	utils.RespondJSON(w, http.StatusOK, nil)
}

func (*authHandler) setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteStrictMode,
	})
}
