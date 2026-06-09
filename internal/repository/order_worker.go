package repository

import "context"

// PendingOrder представляет модель заказа, ожидающего обработки
// во внешней системе начислений.
type PendingOrder struct {
	// Number — номер заказа.
	Number string
	// UserID — идентификатор пользователя, загрузившего заказ.
	UserID int64
}

// FetchPendingOrders получает из базы данных пачку заказов, ожидающих обработки
func (r *OrderRepo) FetchPendingOrders(ctx context.Context, limit int) ([]PendingOrder, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT number, user_id FROM orders
		 WHERE status IN ('NEW', 'PROCESSING')
		 ORDER BY uploaded_at ASC
		 LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []PendingOrder
	for rows.Next() {
		var o PendingOrder
		if err := rows.Scan(&o.Number, &o.UserID); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// UpdateOrderStatus обновляет статус заказа, сумму начисления и время последней
// проверки во внешней системе начислений.
func (r *OrderRepo) UpdateOrderStatus(ctx context.Context, number, status string, accrual *float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $2, accrual = $3, last_checked_at = NOW()
		 WHERE number = $1`, number, status, accrual)
	return err
}
