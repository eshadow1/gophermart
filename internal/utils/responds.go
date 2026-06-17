package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eshadow1/gophermart/internal/loggers"
)

// RespondJSON формирует и отправляет HTTP-ответ в формате JSON.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	errEncode := json.NewEncoder(w).Encode(data)
	if errEncode != nil {
		loggers.Log.Error("failed to encode response", "err", errEncode)
	}
}

// ValidationError представляет ошибку валидации запроса
// (некорректные данные, отсутствующие поля).
type ValidationError struct{ Msg string }

// Error возвращает текстовое представление ошибки валидации.
func (e *ValidationError) Error() string { return e.Msg }

// ConflictError представляет ошибку конфликта
// (например, заказ уже загружен другим пользователем).
type ConflictError struct{ Msg string }

// Error возвращает текстовое представление ошибки валидации.
func (e *ConflictError) Error() string { return e.Msg }

// UnauthorizedError представляет ошибку авторизации
// (отсутствие или невалидность токена).
type UnauthorizedError struct{ Msg string }

// Error возвращает текстовое представление ошибки валидации.
func (e *UnauthorizedError) Error() string { return e.Msg }

// InsufficientFundsError представляет ошибку нехватки средств на балансе.
type InsufficientFundsError struct{ Msg string }

// Error возвращает текстовое представление ошибки валидации.
func (e *InsufficientFundsError) Error() string { return e.Msg }

// IncorrectFormatError представляет ошибку формата данных
// (неверный формат номера заказа, невалидный JSON).
type IncorrectFormatError struct{ Msg string }

// Error возвращает текстовое представление ошибки валидации.
func (e *IncorrectFormatError) Error() string { return e.Msg }

// RespondError анализирует тип ошибки и формирует HTTP-ответ с соответствующим
// статус-кодом и сообщением.
func RespondError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := "internal server error"

	if e, ok := errors.AsType[*ValidationError](err); ok {
		status = http.StatusBadRequest
		msg = e.Msg
	} else if e, ok := errors.AsType[*IncorrectFormatError](err); ok {
		status = http.StatusUnprocessableEntity
		msg = e.Msg
	} else if e, ok := errors.AsType[*ConflictError](err); ok {
		status = http.StatusConflict
		msg = e.Msg
	} else if e, ok := errors.AsType[*UnauthorizedError](err); ok {
		status = http.StatusUnauthorized
		msg = e.Msg
		if e.Msg == "" {
			msg = "unauthorized"
		}
	} else if e, ok := errors.AsType[*InsufficientFundsError](err); ok {
		status = http.StatusPaymentRequired
		msg = e.Msg
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errEncode := json.NewEncoder(w).Encode(map[string]string{"error": msg})
	if errEncode != nil {
		loggers.Log.Error("failed to respond with error", "error", errEncode)
	}
}
