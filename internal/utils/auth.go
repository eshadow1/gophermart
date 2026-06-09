// Package utils предоставляет вспомогательные утилиты, используемые
// в различных слоях приложения (handler, middleware, service).
package utils

import (
	"context"

	"github.com/eshadow1/gophermart/internal/models"
)

// GetUserID извлекает идентификатор пользователя из контекста HTTP-запроса.
// Идентификатор должен быть сохранён в контексте middleware аутентификации
// по ключу models.UserIDContextKey.
func GetUserID(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(models.UserIDContextKey).(int64)
	return uid, ok
}
