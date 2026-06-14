// Package repository предоставляет реализацию слоя доступа к данным (Data Access Layer)
// для работы с базой данных PostgreSQL.
package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/eshadow1/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
)

// BalanceRepo реализует операции для работы с балансами пользователей
type BalanceRepo struct {
	pool DBPool
}

// NewBalanceRepo создает и возвращает новый экземпляр BalanceRepo,
// внедряя зависимость пула соединений с базой данных.
func NewBalanceRepo(pool DBPool) *BalanceRepo {
	return &BalanceRepo{pool: pool}
}

// Get возвращает текущий баланс пользователя по его идентификатору.
func (r *BalanceRepo) Get(ctx context.Context, userID int64) (models.BalanceResponse, error) {
	var res models.BalanceResponse
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(current, 0), COALESCE(withdrawn, 0) 
		 FROM user_balances WHERE user_id = $1`, userID).
		Scan(&res.Current, &res.Withdrawn)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.BalanceResponse{}, nil // Баланс ещё не создан
	}
	return res, err
}

// Withdraw выполняет операцию вывода средств со счета пользователя
// в пользу указанного заказа.
func (r *BalanceRepo) Withdraw(ctx context.Context, userID int64, infoReq models.WithdrawRequest) error {
	tx, errBegin := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if errBegin != nil {
		return errBegin
	}
	defer func() {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			if !errors.Is(errRollback, pgx.ErrTxClosed) {
				slog.Error("transaction rollback failed", "err", errRollback, "tx", tx)
			}
		}
	}()

	res, errUpdate := tx.Exec(ctx,
		`UPDATE user_balances 
		 SET current = current - $2, withdrawn = withdrawn + $2 
		 WHERE user_id = $1 AND current >= $2`,
		userID, infoReq.Sum)
	if errUpdate != nil {
		return errUpdate
	}
	if res.RowsAffected() == 0 {
		return ErrInsufficient
	}

	_, errInsert := tx.Exec(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum, processed_at) 
		 VALUES ($1, $2, $3, NOW())`,
		userID, infoReq.Order, infoReq.Sum)
	if errInsert != nil {
		return errInsert
	}

	return tx.Commit(ctx)
}

// ListWithdrawals возвращает историю всех выводов средств пользователя,
// отсортированную по дате обработки (от новых к старым).
func (r *BalanceRepo) ListWithdrawals(ctx context.Context, userID int64) ([]models.WithdrawalResponse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT order_number, sum, processed_at 
		 FROM withdrawals 
		 WHERE user_id = $1 
		 ORDER BY processed_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.WithdrawalResponse
	for rows.Next() {
		var w models.WithdrawalResponse
		if errScan := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); errScan != nil {
			return nil, errScan
		}
		list = append(list, w)
	}
	return list, rows.Err()
}
