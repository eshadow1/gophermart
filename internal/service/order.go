package service

import (
	"context"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/utils"
)

// OrderRepo определяет контракт для слоя репозитория, работающего
// с заказами пользователей.
type OrderRepo interface {
	// AddOrder добавляет новый заказ в систему от имени пользователя.
	AddOrder(ctx context.Context, userID int64, rawBody string) (bool, error)
	// ListByUser возвращает список всех заказов пользователя.
	ListByUser(ctx context.Context, userID int64) ([]models.OrderResponse, error)
}

// orderService реализует бизнес-логику для работы с заказами пользователей
type orderService struct {
	cfg  *configs.Config
	repo OrderRepo
}

// NewOrderService создает и возвращает новый экземпляр сервиса заказов
// с внедрёнными зависимостями конфигурации и репозитория.
func NewOrderService(cfg *configs.Config, repo OrderRepo) *orderService {
	return &orderService{
		cfg:  cfg,
		repo: repo,
	}
}

// LoadUserOrder загружает номер заказа в систему от имени пользователя.
func (os *orderService) LoadUserOrder(ctx context.Context, userID int64, rawBody string) (bool, error) {
	if rawBody == "" {
		return false, ErrEmptyBody
	}
	if !utils.ValidateLuhn(rawBody) {
		return false, ErrValidationLuhn
	}

	return os.repo.AddOrder(ctx, userID, rawBody)
}

// GetUserOrders возвращает список всех заказов пользователя с их статусами
// и суммами начислений, отсортированный по дате загрузки.
func (os *orderService) GetUserOrders(ctx context.Context, userID int64) ([]models.OrderResponse, error) {
	return os.repo.ListByUser(ctx, userID)
}
