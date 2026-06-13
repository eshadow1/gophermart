package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBPool описывает интерфейс пула соединений с базой данных,
// необходимый для работы BalanceRepo, OrderRepo, UserRepo.
type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	// QueryRow выполняет запрос
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	// Query выполняет запрос
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	// BeginTx начинает новую транзакцию с указанными опциями изоляции.
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}
