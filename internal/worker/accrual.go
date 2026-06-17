// Package worker предоставляет реализацию фоновых процессов (воркеров)
// для асинхронной обработки задач. В данном файле описан воркер для обработки
// заказов через внешнюю систему начислени
package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eshadow1/gophermart/internal/client"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/repository"
	"golang.org/x/sync/errgroup"
)

const (
	statusProcessing = "PROCESSING"
	statusProcessed  = "PROCESSED"
	statusInvalid    = "INVALID"
	maxRetries       = 5
)

// OrderRepo определяет контракт для слоя репозитория, работающего
// с очередью заказов.
type OrderRepo interface {
	// FetchPendingOrders получает пачку заказов, ожидающих обработки
	FetchPendingOrders(ctx context.Context, limit int) ([]repository.PendingOrder, error)
	// UpdateOrderStatus обновляет статус заказа, сумму начисления
	// и время последней проверки во внешней системе.
	UpdateOrderStatus(ctx context.Context, number, status string, accrual *float64) error
}

// BalanceRepo определяет контракт для слоя репозитория, работающего
// с балансами пользователей.
type BalanceRepo interface {
	// AddAccrual начисляет указанное количество баллов на баланс пользователя.
	AddAccrual(ctx context.Context, userID int64, amount float64) error
}

// AccrualWorker реализует фоновый воркер для обработки заказов через
// внешнюю систему начислений.
type AccrualWorker struct {
	orderRepo     OrderRepo
	balanceRepo   BalanceRepo
	accClient     *client.AccrualClient
	batchSize     int
	interval      time.Duration
	maxConcurrent int
}

// NewAccrualWorker создает и возвращает новый экземпляр воркера для обработки
// заказов с внедрёнными зависимостями и настройками конкурентности.
func NewAccrualWorker(orderRepo OrderRepo, balanceRepo BalanceRepo,
	accClient *client.AccrualClient, maxConcurrent int, batchSize int, interval time.Duration) *AccrualWorker {
	return &AccrualWorker{
		orderRepo:     orderRepo,
		balanceRepo:   balanceRepo,
		accClient:     accClient,
		batchSize:     batchSize,
		interval:      interval,
		maxConcurrent: maxConcurrent,
	}
}

// Run запускает основной цикл воркера. Работает в бесконечном цикле,
// обрабатывая пачки заказов с заданным интервалом. Корректно останавливается
// при отмене контекста (graceful shutdown).
func (w *AccrualWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	loggers.Log.Info("accrual worker started", "interval", w.interval)

	for {
		select {
		case <-ctx.Done():
			loggers.Log.Info("accrual worker stopping")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				loggers.Log.Info("batch processing failed", "err", err)
			}
		}
	}
}

// processBatch получает пачку заказов из базы данных и запускает
// их параллельную обработку в отдельных горутинах.
func (w *AccrualWorker) processBatch(ctx context.Context) error {
	orders, err := w.orderRepo.FetchPendingOrders(ctx, w.batchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch pending orders: %w", err)
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(w.maxConcurrent)

	for _, order := range orders {
		ord := order
		g.Go(func() error {
			return w.processOrder(gctx, ord)
		})
	}

	return g.Wait()
}

// processOrder обрабатывает один заказ: запрашивает информацию о начислении
// во внешней системе, обновляет статус заказа и начисляет баллы на баланс.
func (w *AccrualWorker) processOrder(ctx context.Context, o repository.PendingOrder) error {
	for retry := 0; retry <= maxRetries; retry++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		res, err := w.accClient.GetAccrual(ctx, o.Number)
		if err != nil {
			if rl, ok := errors.AsType[*client.RateLimitError](err); ok {
				loggers.Log.Info("rate limited, waiting", "order", o.Number, "delay", rl.RetryAfter)
				select {
				case <-ctx.Done():
					return fmt.Errorf("order %s: context canceled during retry wait: %w", o.Number, ctx.Err())
				case <-time.After(rl.RetryAfter):
					continue
				}
			}
			if retry < maxRetries {
				loggers.Log.Warn("external service error, retrying", "order", o.Number, "err", err)
				continue
			}

			loggers.Log.Warn("external service error, will retry later", "order", o.Number, "err", err)
			return nil
		}

		if res == nil {
			errUpdateStatus := w.orderRepo.UpdateOrderStatus(ctx, o.Number, "INVALID", nil)
			if errUpdateStatus != nil {
				return fmt.Errorf("order %s: failed to add accrual to balance: %w", o.Number, err)
			}
			return nil
		}

		var status string
		var accrual *float64
		switch res.Status {
		case "PROCESSED":
			status = statusProcessed
			accrual = &res.Accrual
			if *accrual > 0 {
				if errBalance := w.balanceRepo.AddAccrual(ctx, o.UserID, *accrual); errBalance != nil {
					return fmt.Errorf("order %s: failed to add accrual to balance: %w", o.Number, errBalance)
				}
			}
		case "INVALID":
			status = statusInvalid
		default:
			status = statusProcessing
		}
		errUpdateStatus := w.orderRepo.UpdateOrderStatus(ctx, o.Number, status, accrual)
		if errUpdateStatus != nil {
			return fmt.Errorf("update order %s status: %w", o.Number, errUpdateStatus)
		}
		return nil
	}

	loggers.Log.Warn("order processing exhausted retries without final status", "order", o.Number)
	return nil
}
