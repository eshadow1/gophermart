package service

import (
	"errors"
	"testing"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/repository"
	mockservice "github.com/eshadow1/gophermart/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBalanceService_GetBalance(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name               string
		userID             int64
		balance            models.BalanceResponse
		errBalance         error
		expectedBalance    models.BalanceResponse
		expectedErrBalance error
	}{
		{
			name:               "success",
			userID:             1,
			balance:            models.BalanceResponse{},
			errBalance:         nil,
			expectedBalance:    models.BalanceResponse{},
			expectedErrBalance: nil,
		},
		{
			name:               "error",
			userID:             1,
			balance:            models.BalanceResponse{},
			errBalance:         errors.New("some error"),
			expectedBalance:    models.BalanceResponse{},
			expectedErrBalance: errors.New("some error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mas := mockservice.NewMockBalanceRepo(t)
			mas.On("Get", t.Context(), mock.Anything, mock.Anything).Return(test.balance, test.errBalance).Maybe()
			h := NewBalanceService(&cfg, mas)

			balance, errReg := h.GetBalance(t.Context(), test.userID)

			assert.Equal(t, test.expectedBalance, balance)
			if test.expectedErrBalance != nil {
				assert.Equal(t, test.expectedErrBalance, errReg)
			} else {
				assert.NoError(t, errReg)
			}
		})
	}
}

func TestBalanceService_ListWithdrawals(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name               string
		userID             int64
		balance            []models.WithdrawalResponse
		errBalance         error
		expectedBalance    []models.WithdrawalResponse
		expectedErrBalance error
	}{
		{
			name:               "success",
			userID:             1,
			balance:            make([]models.WithdrawalResponse, 1),
			errBalance:         nil,
			expectedBalance:    make([]models.WithdrawalResponse, 1),
			expectedErrBalance: nil,
		},
		{
			name:               "error",
			userID:             1,
			balance:            make([]models.WithdrawalResponse, 0),
			errBalance:         errors.New("some error"),
			expectedBalance:    make([]models.WithdrawalResponse, 0),
			expectedErrBalance: errors.New("some error"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mas := mockservice.NewMockBalanceRepo(t)
			mas.On("ListWithdrawals", t.Context(), mock.Anything, mock.Anything).Return(test.balance, test.errBalance).Maybe()
			h := NewBalanceService(&cfg, mas)

			balance, errReg := h.ListWithdrawals(t.Context(), test.userID)

			assert.Equal(t, test.expectedBalance, balance)
			if test.expectedErrBalance != nil {
				assert.Equal(t, test.expectedErrBalance, errReg)
			} else {
				assert.NoError(t, errReg)
			}
		})
	}
}

func TestBalanceService_RequestWithdraw(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name               string
		userID             int64
		withdraw           models.WithdrawRequest
		errWithdraw        error
		expectedErrRequest error
	}{
		{
			name:   "success",
			userID: 1,
			withdraw: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   0.2,
			},
			errWithdraw:        nil,
			expectedErrRequest: nil,
		},
		{
			name:   "error_validation_luhn",
			userID: 1,
			withdraw: models.WithdrawRequest{
				Order: "",
				Sum:   0.2,
			},
			errWithdraw:        nil,
			expectedErrRequest: ErrValidationLuhn,
		},
		{
			name:   "error_sum_positive",
			userID: 1,
			withdraw: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   0,
			},
			errWithdraw:        nil,
			expectedErrRequest: errors.New("withdrawal sum must be positive"),
		},
		{
			name:   "error_insufficient",
			userID: 1,
			withdraw: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   0.2,
			},
			errWithdraw:        repository.ErrInsufficient,
			expectedErrRequest: ErrInsufficient,
		},
		{
			name:   "error_order_not_found",
			userID: 1,
			withdraw: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   0.2,
			},
			errWithdraw:        repository.ErrOrder,
			expectedErrRequest: ErrOrderNotFound,
		},
		{
			name:   "error_server",
			userID: 1,
			withdraw: models.WithdrawRequest{
				Order: "12345678903",
				Sum:   0.2,
			},
			errWithdraw:        errors.New("some error"),
			expectedErrRequest: errors.New("some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mas := mockservice.NewMockBalanceRepo(t)
			mas.On("Withdraw", t.Context(), mock.Anything, mock.Anything).Return(test.errWithdraw).Maybe()
			h := NewBalanceService(&cfg, mas)

			errReg := h.RequestWithdraw(t.Context(), test.userID, test.withdraw)

			if test.expectedErrRequest != nil {
				assert.Equal(t, test.expectedErrRequest, errReg)
			} else {
				assert.NoError(t, errReg)
			}
		})
	}
}
