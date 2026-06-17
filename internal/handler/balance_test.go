package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/models"
	"github.com/eshadow1/gophermart/internal/service"
	mockhandler "github.com/eshadow1/gophermart/mocks/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBalanceHandler_GetBalanceUser(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name                string
		method              string
		userID              int64
		balance             models.BalanceResponse
		errBalance          error
		expectedContentType string
		expectedBody        string
		expectedStatus      int
	}{
		{
			name:   "success",
			method: http.MethodGet,
			userID: 1,
			balance: models.BalanceResponse{
				Current:   1.0,
				Withdrawn: 1.0,
			},
			errBalance:          nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusOK,
			expectedBody:        "{\"current\":1,\"withdrawn\":1}\n",
		},
		{
			name:                "unauthorized",
			method:              http.MethodGet,
			userID:              0,
			balance:             models.BalanceResponse{},
			errBalance:          nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusUnauthorized,
			expectedBody:        "{\"error\":\"unauthorized\"}\n",
		},
		{
			name:                "error_get_user_balance",
			method:              http.MethodGet,
			userID:              1,
			balance:             models.BalanceResponse{},
			errBalance:          errors.New("internal server erro"),
			expectedContentType: "application/json",
			expectedStatus:      http.StatusInternalServerError,
			expectedBody:        "{\"error\":\"internal server error\"}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/", http.NoBody)
			req.Header.Set("Content-Type", "application/json")

			if test.userID != 0 {
				ctx := context.WithValue(t.Context(), models.UserIDContextKey, test.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			mbs := mockhandler.NewMockBalanceServer(t)
			mbs.On("GetBalance", req.Context(), mock.Anything).Return(test.balance, test.errBalance).Maybe()
			h := NewBalanceHandler(&cfg, mbs)

			h.GetBalanceUser(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

func TestBalanceHandler_GetInfoWithdrawals(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name                string
		method              string
		userID              int64
		balance             []models.WithdrawalResponse
		errBalance          error
		expectedContentType string
		expectedBody        string
		expectedStatus      int
	}{
		{
			name:   "success",
			method: http.MethodGet,
			userID: 1,
			balance: []models.WithdrawalResponse{
				{
					Order:       "1",
					Sum:         1.0,
					ProcessedAt: time.Time{},
				},
			},
			errBalance:          nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusOK,
			expectedBody:        "[{\"order\":\"1\",\"sum\":1,\"processed_at\":\"0001-01-01T00:00:00Z\"}]\n",
		},
		{
			name:                "success_without_balance",
			method:              http.MethodGet,
			userID:              1,
			balance:             []models.WithdrawalResponse{},
			errBalance:          nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusNoContent,
			expectedBody:        "",
		},
		{
			name:   "unauthorized",
			method: http.MethodGet,
			userID: 0,
			balance: []models.WithdrawalResponse{
				{
					Order:       "1",
					Sum:         1.0,
					ProcessedAt: time.Time{},
				},
			},
			errBalance:          nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusUnauthorized,
			expectedBody:        "{\"error\":\"unauthorized\"}\n",
		},
		{
			name:   "error_get_user_balance",
			method: http.MethodGet,
			userID: 1,
			balance: []models.WithdrawalResponse{
				{
					Order:       "1",
					Sum:         1.0,
					ProcessedAt: time.Time{},
				},
			},
			errBalance:          errors.New("internal server error"),
			expectedContentType: "application/json",
			expectedStatus:      http.StatusInternalServerError,
			expectedBody:        "{\"error\":\"internal server error\"}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/", http.NoBody)
			req.Header.Set("Content-Type", "application/json")

			if test.userID != 0 {
				ctx := context.WithValue(t.Context(), models.UserIDContextKey, test.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			mbs := mockhandler.NewMockBalanceServer(t)
			mbs.On("ListWithdrawals", req.Context(), mock.Anything).Return(test.balance, test.errBalance).Maybe()
			h := NewBalanceHandler(&cfg, mbs)

			h.GetInfoWithdrawals(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

func TestBalanceHandler_RequestWithdraw(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name                string
		method              string
		userID              int64
		body                string
		errWithdraw         error
		expectedContentType string
		expectedBody        string
		expectedStatus      int
	}{
		{
			name:                "success",
			method:              http.MethodPost,
			userID:              1,
			body:                "{\"order\": \"2377225624\",\"sum\": 751\n}",
			errWithdraw:         nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusOK,
			expectedBody:        "",
		},
		{
			name:                "incorrect_body",
			method:              http.MethodGet,
			userID:              1,
			body:                "{\"order\": \"2377225624\",\"sum\": 751,\n}",
			errWithdraw:         nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        "{\"error\":\"invalid json format\"}\n",
		},
		{
			name:                "unauthorized",
			method:              http.MethodGet,
			userID:              0,
			body:                "{\"order\": \"2377225624\",\"sum\": 751\n}",
			errWithdraw:         nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusUnauthorized,
			expectedBody:        "{\"error\":\"unauthorized\"}\n",
		},
		{
			name:                "error_internal_server_error",
			method:              http.MethodGet,
			userID:              1,
			body:                "{\"order\": \"2377225624\",\"sum\": 751\n}",
			errWithdraw:         errors.New("internal server error"),
			expectedContentType: "application/json",
			expectedStatus:      http.StatusInternalServerError,
			expectedBody:        "{\"error\":\"internal server error\"}\n",
		},
		{
			name:                "error_internal_server_error",
			method:              http.MethodGet,
			userID:              1,
			body:                "{\"order\": \"2377225624\",\"sum\": 751\n}",
			errWithdraw:         service.ErrOrderNotFound,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusUnprocessableEntity,
			expectedBody:        "{\"error\":\"order not found\"}\n",
		},
		{
			name:                "error_internal_server_error",
			method:              http.MethodGet,
			userID:              1,
			body:                "{\"order\": \"2377225624\",\"sum\": 751\n}",
			errWithdraw:         service.ErrInsufficient,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusPaymentRequired,
			expectedBody:        "{\"error\":\"insufficient funds\"}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")

			if test.userID != 0 {
				ctx := context.WithValue(t.Context(), models.UserIDContextKey, test.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			mbs := mockhandler.NewMockBalanceServer(t)
			mbs.On("RequestWithdraw", req.Context(), mock.Anything, mock.Anything).Return(test.errWithdraw).Maybe()
			h := NewBalanceHandler(&cfg, mbs)

			h.RequestWithdraw(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}
