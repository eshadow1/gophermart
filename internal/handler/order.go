package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/repository"
	"github.com/eshadow1/gophermart/internal/service"
	"github.com/eshadow1/gophermart/internal/utils"
)

// OrderServer определяет контракт для слоя бизнес-логики,
// отвечающего за обработку и хранение заказов пользователей.
type OrderServer interface {
	// LoadUserOrder загружает номер заказа в систему от имени пользователя.
	LoadUserOrder(context.Context, int64, string) (bool, error)
	// GetUserOrders возвращает список всех загруженных заказов пользователя
	// с информацией об их статусах и суммах начислений.
	GetUserOrders(context.Context, int64) ([]models.OrderResponse, error)
}

// orderHandler реализует HTTP-хендлеры для эндпоинтов загрузки и получения заказов.
type orderHandler struct {
	cfg *configs.Config
	s   OrderServer
}

// NewOrderHandler создает и возвращает новый экземпляр HTTP-хендлера для работы с заказами.
// Принимает конфигурацию приложения и реализацию интерфейса OrderServer.
func NewOrderHandler(cfg *configs.Config, svc OrderServer) *orderHandler {
	return &orderHandler{
		cfg: cfg,
		s:   svc,
	}
}

// UploadUserOrder обрабатывает HTTP-запросы на загрузку номера заказа.
func (h *orderHandler) UploadUserOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, &utils.UnauthorizedError{})
		return
	}
	if r.Header.Get("Content-Type") != "text/plain" {
		utils.RespondError(w, &utils.ValidationError{Msg: "invalid request content-type"})
		return
	}

	body, errRead := io.ReadAll(r.Body)
	if errRead != nil || strings.TrimSpace(string(body)) == "" {
		utils.RespondError(w, &utils.ValidationError{Msg: "invalid request body"})
		return
	}
	loggers.Log.Infow("order request", "id", userID, "body", string(body))

	isInsert, err := h.s.LoadUserOrder(r.Context(), userID, strings.TrimSpace(string(body)))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidationLuhn):
			utils.RespondError(w, &utils.IncorrectFormatError{Msg: err.Error()})
		case errors.Is(err, service.ErrEmptyBody):
			utils.RespondError(w, &utils.ValidationError{Msg: err.Error()})
		case errors.Is(err, repository.ErrConflict):
			utils.RespondError(w, &utils.ConflictError{Msg: err.Error()})
		default:
			utils.RespondError(w, err)
		}
		return
	}

	if isInsert {
		utils.RespondJSON(w, http.StatusAccepted, nil)
		return
	}
	utils.RespondJSON(w, http.StatusOK, nil)
}

// GetUserOrders обрабатывает HTTP-запросы на получение списка заказов пользователя.
func (h *orderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, &utils.UnauthorizedError{})
		return
	}

	orders, err := h.s.GetUserOrders(r.Context(), userID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	if len(orders) == 0 {
		utils.RespondJSON(w, http.StatusNoContent, nil)
		return
	}
	utils.RespondJSON(w, http.StatusOK, orders)
}
