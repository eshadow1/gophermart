package service

import (
	"errors"
	"testing"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/models"
	mockservice "github.com/eshadow1/gophermart/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Register(t *testing.T) {
	cfg := configs.AuthConfig{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name           string
		user           *models.User
		errLogin       error
		errCreate      error
		login          string
		password       string
		expectedToken  string
		expectedErrReg error
	}{
		{
			name: "success",
			user: &models.User{
				ID:           1,
				Login:        "login",
				PasswordHash: "login123",
			},
			errLogin:       nil,
			errCreate:      nil,
			login:          "login",
			password:       "login123",
			expectedToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedErrReg: nil,
		},
		{
			name: "incorrect_password",
			user: &models.User{
				ID:           1,
				Login:        "login",
				PasswordHash: "login",
			},
			errLogin:       nil,
			errCreate:      nil,
			login:          "login",
			password:       "login",
			expectedToken:  "",
			expectedErrReg: models.ErrInvalidFormat,
		},
		{
			name: "error_create",
			user: &models.User{
				ID:           1,
				Login:        "login",
				PasswordHash: "login123",
			},
			errLogin:       nil,
			errCreate:      errors.New("error"),
			login:          "login",
			password:       "login123",
			expectedToken:  "",
			expectedErrReg: errors.New("error"),
		},
		{
			name: "error_login",
			user: &models.User{
				ID:           1,
				Login:        "login",
				PasswordHash: "login123",
			},
			errLogin:       errors.New("error"),
			errCreate:      nil,
			login:          "login",
			password:       "login123",
			expectedToken:  "",
			expectedErrReg: errors.New("error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mas := mockservice.NewMockUserRepo(t)
			mas.On("GetByLogin", t.Context(), mock.Anything, mock.Anything).Return(test.user, test.errLogin).Maybe()
			mas.On("Create", t.Context(), mock.Anything, mock.Anything).Return(test.errCreate).Maybe()
			h := NewAuthService(&cfg, mas)

			token, errReg := h.Register(t.Context(), test.login, test.password)

			assert.Contains(t, token, test.expectedToken)
			if test.expectedErrReg != nil {
				assert.Equal(t, test.expectedErrReg, errReg)
			} else {
				assert.NoError(t, errReg)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	cfg := configs.AuthConfig{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name           string
		user           *models.User
		errLogin       error
		login          string
		password       string
		expectedToken  string
		expectedErrReg error
	}{
		{
			name: "success",
			user: &models.User{
				ID:           1,
				Login:        "login",
				PasswordHash: "$2a$10$rKUsVWrctDwJ45LgOR8DOu55JAJeOHyrbRD2LK3Adk1BnafDc.O02",
			},
			errLogin:       nil,
			login:          "login",
			password:       "login123",
			expectedToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedErrReg: nil,
		},
		{
			name: "incorrect_password",
			user: &models.User{
				ID:           1,
				Login:        "login",
				PasswordHash: "$2a$10$rKUsVWrctDwJ45LgOR8DOu55JAJeOHyrbRD2LK3Adk1BnafDc",
			},
			errLogin:       nil,
			login:          "login",
			password:       "incorrect123",
			expectedToken:  "",
			expectedErrReg: models.ErrInvalidCredentials,
		},
		{
			name:           "incorrect_login",
			user:           nil,
			errLogin:       errors.New("error"),
			login:          "login",
			password:       "login123",
			expectedToken:  "",
			expectedErrReg: errors.New("error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mas := mockservice.NewMockUserRepo(t)
			mas.On("GetByLogin", t.Context(), mock.Anything, mock.Anything).Return(test.user, test.errLogin).Maybe()
			h := NewAuthService(&cfg, mas)

			token, errReg := h.Login(t.Context(), test.login, test.password)

			assert.Contains(t, token, test.expectedToken)
			if test.expectedErrReg != nil {
				assert.Equal(t, test.expectedErrReg, errReg)
			} else {
				assert.NoError(t, errReg)
			}
		})
	}
}
