package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/service"
	"github.com/eshadow1/gophermart/internal/utils"
)

// BalanceServer определяет контракт для слоя бизнес-логики,
// отвечающего за управление балансом пользователей и обработку выводов средств.
type BalanceServer interface {
	// GetBalance возвращает текущий баланс и сумму заблокированных средств пользователя.
	GetBalance(context.Context, int64) (models.BalanceResponse, error)
	// RequestWithdraw инициирует процесс снятия средств со счета пользователя
	// в пользу указанного заказа.
	RequestWithdraw(context.Context, int64, models.WithdrawRequest) error
	// ListWithdrawals возвращает историю всех успешных выводов средств пользователя.
	ListWithdrawals(context.Context, int64) ([]models.WithdrawalResponse, error)
}

// balanceHandler реализует HTTP-хендлеры для эндпоинтов управления балансом и выводами.
type balanceHandler struct {
	cfg *configs.Config
	s   BalanceServer
}

// NewBalanceHandler создает и возвращает новый экземпляр HTTP-хендлера для работы с балансом.
// Принимает конфигурацию приложения и реализацию интерфейса BalanceServer.
func NewBalanceHandler(cfg *configs.Config, svc BalanceServer) *balanceHandler {
	return &balanceHandler{
		cfg: cfg,
		s:   svc,
	}
}

// GetBalanceUser обрабатывает HTTP-запросы на получение текущего баланса пользователя.
func (h *balanceHandler) GetBalanceUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, &utils.UnauthorizedError{})
		return
	}

	bal, err := h.s.GetBalance(r.Context(), userID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}
	utils.RespondJSON(w, http.StatusOK, bal)
}

// RequestWithdraw обрабатывает HTTP-запросы на снятие средств (вывод) со счета пользователя.
// Ожидает JSON с данными для вывода.
func (h *balanceHandler) RequestWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, &utils.UnauthorizedError{})
		return
	}

	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, &utils.ValidationError{Msg: "invalid json format"})
		return
	}

	loggers.Log.Infow("order request", "id", userID, "body", req)

	err := h.s.RequestWithdraw(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			utils.RespondError(w, &utils.IncorrectFormatError{Msg: err.Error()})
		case errors.Is(err, service.ErrInsufficient):
			utils.RespondError(w, &utils.InsufficientFundsError{Msg: err.Error()})
		default:
			utils.RespondError(w, err)
		}
		return
	}

	utils.RespondJSON(w, http.StatusOK, nil)
}

// GetInfoWithdrawals обрабатывает HTTP-запросы на получение истории выводов средств.
func (h *balanceHandler) GetInfoWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r.Context())
	if !ok {
		utils.RespondError(w, &utils.UnauthorizedError{})
		return
	}

	list, err := h.s.ListWithdrawals(r.Context(), userID)
	if err != nil {
		utils.RespondError(w, err)
		return
	}

	if len(list) == 0 {
		utils.RespondJSON(w, http.StatusNoContent, nil)
		return
	}
	utils.RespondJSON(w, http.StatusOK, list)
}
