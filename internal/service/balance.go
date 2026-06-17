package service

import (
	"context"
	"errors"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/repository"
	"github.com/eshadow1/gophermart/internal/utils"
)

// BalanceRepo определяет контракт для слоя репозитория, работающего
// с балансами пользователей и операциями вывода средств.
type BalanceRepo interface {
	// Get возвращает текущий баланс пользователя по его идентификатору.
	Get(context.Context, int64) (models.BalanceResponse, error)
	// Withdraw выполняет операцию вывода средств со счета пользователя.
	Withdraw(context.Context, int64, models.WithdrawRequest) error
	// ListWithdrawals возвращает историю всех выводов средств пользователя.
	ListWithdrawals(context.Context, int64) ([]models.WithdrawalResponse, error)
}

// BalanceService реализует бизнес-логику для работы с балансом пользователей.
type BalanceService struct {
	repo BalanceRepo
}

// NewBalanceService создает и возвращает новый экземпляр сервиса баланса
// с внедрённой зависимостью репозитория.
func NewBalanceService(_ *configs.Config, repo BalanceRepo) *BalanceService {
	return &BalanceService{repo: repo}
}

// GetBalance возвращает текущий баланс пользователя, включая сумму
// доступных средств и общую сумму выведенных средств.
func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (models.BalanceResponse, error) {
	return s.repo.Get(ctx, userID)
}

// RequestWithdraw обрабатывает запрос на вывод средств со счета пользователя.
func (s *BalanceService) RequestWithdraw(ctx context.Context, userID int64, req models.WithdrawRequest) error {
	if !utils.ValidateLuhn(req.Order) {
		return ErrValidationLuhn
	}
	if req.Sum <= 0 {
		return errors.New("withdrawal sum must be positive")
	}

	err := s.repo.Withdraw(ctx, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInsufficient):
			return ErrInsufficient
		case errors.Is(err, repository.ErrOrder):
			return ErrOrderNotFound
		default:
			return err
		}
	}
	return nil
}

// ListWithdrawals возвращает историю всех успешных выводов средств пользователя,
// отсортированную по дате обработки.
func (s *BalanceService) ListWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalResponse, error) {
	return s.repo.ListWithdrawals(ctx, userID)
}
