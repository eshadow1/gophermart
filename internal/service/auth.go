// Package service предоставляет реализацию бизнес-логики приложения.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserRepo определяет контракт для слоя репозитория, работающего
// с учётными записями пользователей.
type UserRepo interface {
	// Create создаёт новую запись пользователя в базе данных.
	Create(ctx context.Context, u *models.User) error
	// GetByLogin возвращает пользователя по логину.
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}

// authService реализует бизнес-логику аутентификации: регистрацию,
// вход в систему и генерацию JWT-токенов.
type authService struct {
	repo UserRepo
	cfg  *configs.AuthConfig
}

// NewAuthService создает и возвращает новый экземпляр сервиса аутентификации
// с внедрёнными зависимостями конфигурации и репозитория пользователей.
func NewAuthService(cfg *configs.AuthConfig, repo UserRepo) *authService {
	return &authService{
		repo: repo,
		cfg:  cfg,
	}
}

const (
	minSizePassword = 6
	lifeToken       = 24 * time.Hour
)

// Register регистрирует нового пользователя в системе.
func (s *authService) Register(ctx context.Context, login, password string) (string, error) {
	if len(password) < minSizePassword {
		return "", models.ErrInvalidFormat
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate hash: %w", err)
	}

	u := &models.User{Login: login, PasswordHash: string(hash)}
	if errCreate := s.repo.Create(ctx, u); errCreate != nil {
		return "", errCreate
	}

	created, errLogin := s.repo.GetByLogin(ctx, login)
	if errLogin != nil {
		return "", errLogin
	}

	return s.generateToken(created.ID)
}

// Login выполняет вход пользователя в систему.
func (s *authService) Login(ctx context.Context, login, password string) (string, error) {
	u, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", models.ErrInvalidCredentials
	}

	return s.generateToken(u.ID)
}

// generateToken создает JWT-токен с указанным userID в поле "sub"
// и сроком действия 24 часа.
func (s *authService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(lifeToken).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.cfg.JWTSecret)
}
