// Package worker предоставляет реализацию фоновых процессов (воркеров)
// для асинхронной обработки задач. В данном файле описан воркер для обработки
// заказов через внешнюю систему начислени
package worker

import (
	"context"
	"time"

	"github.com/eshadow1/gophermart/internal/client"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/repository"
	"golang.org/x/sync/semaphore"
)

const (
	statusProcessing = "PROCESSING"
	statusInvalid    = "INVALID"
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
	orderRepo   OrderRepo
	balanceRepo BalanceRepo
	accClient   *client.AccrualClient
	sem         *semaphore.Weighted
	batchSize   int
	interval    time.Duration
}

// NewAccrualWorker создает и возвращает новый экземпляр воркера для обработки
// заказов с внедрёнными зависимостями и настройками конкурентности.
func NewAccrualWorker(orderRepo OrderRepo, balanceRepo BalanceRepo,
	accClient *client.AccrualClient, maxConcurrent int, batchSize int, interval time.Duration) *AccrualWorker {
	return &AccrualWorker{
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
		accClient:   accClient,
		sem:         semaphore.NewWeighted(int64(maxConcurrent)),
		batchSize:   batchSize,
		interval:    interval,
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
			w.processBatch(ctx)
		}
	}
}

// processBatch получает пачку заказов из базы данных и запускает
// их параллельную обработку в отдельных горутинах.
func (w *AccrualWorker) processBatch(ctx context.Context) {
	orders, err := w.orderRepo.FetchPendingOrders(ctx, w.batchSize)
	if err != nil {
		loggers.Log.Error("failed to fetch pending orders", "err", err)
		return
	}

	for _, o := range orders {
		if err := w.sem.Acquire(ctx, 1); err != nil {
			return // Контекст отменён во время ожидания семафора
		}
		go w.processOrder(ctx, o)
	}
}

// processOrder обрабатывает один заказ: запрашивает информацию о начислении
// во внешней системе, обновляет статус заказа и начисляет баллы на баланс.
func (w *AccrualWorker) processOrder(ctx context.Context, o repository.PendingOrder) {
	defer w.sem.Release(1)

	for {
		res, err := w.accClient.GetAccrual(ctx, o.Number)
		if err != nil {
			if rl, ok := err.(*client.RateLimitError); ok {
				loggers.Log.Info("rate limited, waiting", "order", o.Number, "delay", rl.RetryAfter)
				select {
				case <-ctx.Done():
					return
				case <-time.After(rl.RetryAfter):
					continue
				}
			}
			loggers.Log.Warn("external service error, will retry later", "order", o.Number, "err", err)
			return
		}

		if res == nil {
			errUpdateStatus := w.orderRepo.UpdateOrderStatus(ctx, o.Number, "INVALID", nil)
			if errUpdateStatus != nil {
				loggers.Log.Error("failed to add accrual to balance", "order", o.Number, "err", err)
			}
			return
		}

		var status string
		var accrual *float64
		switch res.Status {
		case "PROCESSED":
			status = statusProcessing
			accrual = &res.Accrual
			if *accrual > 0 {
				if errBalance := w.balanceRepo.AddAccrual(ctx, o.UserID, *accrual); errBalance != nil {
					loggers.Log.Error("failed to add accrual to balance", "order", o.Number, "err", errBalance)
					return
				}
			}
		case "INVALID":
			status = statusInvalid
		default:
			status = statusProcessing
		}
		errUpdateStatus := w.orderRepo.UpdateOrderStatus(ctx, o.Number, status, accrual)
		if errUpdateStatus != nil {
			loggers.Log.Error("failed to add accrual to balance", "order", o.Number, "err", err)
			continue
		}
		return
	}
}
