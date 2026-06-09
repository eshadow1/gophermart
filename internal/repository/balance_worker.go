package repository

import (
	"context"
)

// AddAccrual начисляет указанное количество баллов на баланс пользователя.
func (r *BalanceRepo) AddAccrual(ctx context.Context, userID int64, amount float64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_balances (user_id, current, withdrawn)
		 VALUES ($1, $2, 0)
		 ON CONFLICT (user_id) DO UPDATE 
		 SET current = user_balances.current + EXCLUDED.current`,
		userID, amount)
	return err
}
