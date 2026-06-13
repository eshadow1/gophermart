package repository

import (
	"context"
	"errors"

	"github.com/eshadow1/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	// codePostgresDuplicateInsert — код ошибки PostgreSQL для нарушения UNIQUE-ограничения
	codePostgresDuplicateInsert = "23505"
)

// UserRepo реализует операции для работы с учётными записями пользователей
// в базе данных PostgreSQL
type UserRepo struct {
	pool DBPool
}

// NewUserRepo создает и возвращает новый экземпляр UserRepo,
// внедряя зависимость пула соединений с базой данных.
func NewUserRepo(pool DBPool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create создает новую запись пользователя в базе данных.
func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO users (login, password_hash) VALUES ($1, $2)",
		u.Login, u.PasswordHash,
	)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == codePostgresDuplicateInsert {
			return models.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

// GetByLogin возвращает пользователя по логину для аутентификации.
func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		"SELECT id, login, password_hash FROM users WHERE login = $1", login,
	).Scan(&u.ID, &u.Login, &u.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrInvalidCredentials
		}
		return nil, err
	}
	return &u, nil
}
