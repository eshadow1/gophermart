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

func TestOrderService_GetUserOrders(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name               string
		userID             int64
		orders             []models.OrderResponse
		errOrders          error
		expectedOrders     []models.OrderResponse
		expectedErrBalance error
	}{
		{
			name:               "success",
			userID:             1,
			orders:             make([]models.OrderResponse, 1),
			errOrders:          nil,
			expectedOrders:     make([]models.OrderResponse, 1),
			expectedErrBalance: nil,
		},
		{
			name:               "error",
			userID:             1,
			orders:             make([]models.OrderResponse, 0),
			errOrders:          errors.New("error"),
			expectedOrders:     make([]models.OrderResponse, 0),
			expectedErrBalance: errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mas := mockservice.NewMockOrderRepo(t)
			mas.On("ListByUser", t.Context(), mock.Anything).Return(test.orders, test.errOrders).Maybe()
			h := NewOrderService(&cfg, mas)

			orders, errReg := h.GetUserOrders(t.Context(), test.userID)

			assert.Equal(t, test.expectedOrders, orders)
			if test.expectedErrBalance != nil {
				assert.Equal(t, test.expectedErrBalance, errReg)
			} else {
				assert.NoError(t, errReg)
			}
		})
	}
}

func TestOrderService_LoadUserOrder(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name            string
		userID          int64
		body            string
		isLoad          bool
		errLoad         error
		expectedLoad    bool
		expectedLoadErr error
	}{
		{
			name:            "success",
			userID:          1,
			body:            "12345678903",
			isLoad:          true,
			errLoad:         nil,
			expectedLoad:    true,
			expectedLoadErr: nil,
		},
		{
			name:            "error_empty_body",
			userID:          1,
			body:            "",
			isLoad:          false,
			errLoad:         nil,
			expectedLoad:    false,
			expectedLoadErr: ErrEmptyBody,
		},
		{
			name:            "error_validation_luhn",
			userID:          1,
			body:            "1234567890",
			isLoad:          false,
			errLoad:         nil,
			expectedLoad:    false,
			expectedLoadErr: ErrValidationLuhn,
		},
		{
			name:            "error_add_orde",
			userID:          1,
			body:            "12345678903",
			isLoad:          false,
			errLoad:         errors.New("error"),
			expectedLoad:    false,
			expectedLoadErr: errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mas := mockservice.NewMockOrderRepo(t)
			mas.On("AddOrder", t.Context(), mock.Anything, mock.Anything).Return(test.isLoad, test.errLoad).Maybe()
			h := NewOrderService(&cfg, mas)

			orders, errReg := h.LoadUserOrder(t.Context(), test.userID, test.body)

			assert.Equal(t, test.expectedLoad, orders)
			if test.expectedLoadErr != nil {
				assert.Equal(t, test.expectedLoadErr, errReg)
			} else {
				assert.NoError(t, errReg)
			}
		})
	}
}
