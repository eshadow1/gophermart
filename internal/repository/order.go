// Package repository предоставляет реализацию слоя доступа к данным
// для работы с базой данных PostgreSQL.
package repository

import (
	"context"
	"errors"

	"github.com/eshadow1/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
)

// OrderRepo реализует операции для работы с заказами пользователей
// в базе данных PostgreSQL
type OrderRepo struct {
	pool DBPool
}

// NewOrderRepo создает и возвращает новый экземпляр OrderRepo
func NewOrderRepo(pool DBPool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

// AddOrder добавляет новый заказ в систему от имени пользователя.
func (r *OrderRepo) AddOrder(ctx context.Context, userID int64, number string) (bool, error) {
	var existingUserID int64
	err := r.pool.QueryRow(ctx, "SELECT user_id FROM orders WHERE number = $1", number).Scan(&existingUserID)
	if err == nil {
		if existingUserID == userID {
			return false, nil
		}
		return false, ErrConflict
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	insertQuery := "INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, 'NEW', NOW())"
	_, err = r.pool.Exec(ctx, insertQuery, userID, number)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ListByUser возвращает список всех заказов пользователя, отсортированный
// по дате загрузки (от новых к старым).
func (r *OrderRepo) ListByUser(ctx context.Context, userID int64) ([]models.OrderResponse, error) {
	selectQuery := `SELECT number, status::text, accrual, uploaded_at 
		 FROM orders 
		 WHERE user_id = $1 
		 ORDER BY uploaded_at DESC`

	rows, errSelect := r.pool.Query(ctx, selectQuery, userID)
	if errSelect != nil {
		return nil, errSelect
	}
	defer rows.Close()

	var orders []models.OrderResponse
	for rows.Next() {
		var order models.OrderResponse
		var accrual *float64
		if errScan := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt); errScan != nil {
			return nil, errScan
		}
		if accrual != nil {
			order.Accrual = accrual
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}
