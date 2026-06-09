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
	"github.com/eshadow1/gophermart/internal/repository"
	"github.com/eshadow1/gophermart/internal/service"
	mockhandler "github.com/eshadow1/gophermart/mocks/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderHandler_GetUserOrders(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name                string
		method              string
		userID              int64
		orders              []models.OrderResponse
		errOrder            error
		expectedContentType string
		expectedBody        string
		expectedStatus      int
	}{
		{
			name:                "success_no_content",
			method:              http.MethodGet,
			userID:              1,
			orders:              make([]models.OrderResponse, 0),
			errOrder:            nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusNoContent,
			expectedBody:        ``,
		},
		{
			name:   "success_with_content",
			method: http.MethodPost,
			userID: 1,
			orders: []models.OrderResponse{
				{Number: "1",
					Status:     "NEW",
					UploadedAt: time.Time{},
				},
			},
			errOrder:            nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusOK,
			expectedBody:        "[{\"number\":\"1\",\"status\":\"NEW\",\"uploaded_at\":\"0001-01-01T00:00:00Z\"}]\n",
		},
		{
			name:   "unauthorized",
			method: http.MethodPost,
			userID: 0,
			orders: []models.OrderResponse{
				{Number: "1",
					Status:     "NEW",
					UploadedAt: time.Now(),
				},
			},
			errOrder:            nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusUnauthorized,
			expectedBody:        "{\"error\":\"unauthorized\"}\n",
		},
		{
			name:   "error_get_user_orders",
			method: http.MethodPost,
			userID: 1,
			orders: []models.OrderResponse{
				{Number: "1",
					Status:     "NEW",
					UploadedAt: time.Now(),
				},
			},
			errOrder:            errors.New("internal server erro"),
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

			mos := mockhandler.NewMockOrderServer(t)
			mos.On("GetUserOrders", req.Context(), mock.Anything).Return(test.orders, test.errOrder).Maybe()
			h := NewOrderHandler(&cfg, mos)

			h.GetUserOrders(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}

func TestOrderHandler_UploadUserOrder(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name                string
		method              string
		userID              int64
		contentType         string
		body                string
		isInsert            bool
		errInsert           error
		expectedContentType string
		expectedBody        string
		expectedStatus      int
	}{
		{
			name:                "success_insert",
			method:              http.MethodPost,
			userID:              1,
			contentType:         "text/plain",
			body:                "123456789",
			isInsert:            false,
			errInsert:           nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusOK,
			expectedBody:        ``,
		},
		{
			name:                "success_accepted",
			method:              http.MethodPost,
			userID:              1,
			contentType:         "text/plain",
			body:                "123456789",
			isInsert:            true,
			errInsert:           nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusAccepted,
			expectedBody:        ``,
		},
		{
			name:                "unauthorized",
			method:              http.MethodPost,
			userID:              0,
			contentType:         "text/plain",
			body:                "123456789",
			isInsert:            true,
			errInsert:           nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusUnauthorized,
			expectedBody:        "{\"error\":\"unauthorized\"}\n",
		},
		{
			name:                "invalid_content_type",
			method:              http.MethodPost,
			userID:              1,
			contentType:         "application/json",
			body:                "123456789",
			isInsert:            false,
			errInsert:           nil,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        "{\"error\":\"invalid request content-type\"}\n",
		},
		{
			name:                "validation_luhn",
			method:              http.MethodPost,
			userID:              1,
			contentType:         "text/plain",
			body:                "123456789",
			isInsert:            false,
			errInsert:           service.ErrValidationLuhn,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusUnprocessableEntity,
			expectedBody:        "{\"error\":\"invalid order number format\"}\n",
		},
		{
			name:                "empty_body",
			method:              http.MethodPost,
			userID:              1,
			contentType:         "text/plain",
			body:                "",
			isInsert:            false,
			errInsert:           service.ErrEmptyBody,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusBadRequest,
			expectedBody:        "{\"error\":\"invalid request body\"}\n",
		},
		{
			name:                "conflict",
			method:              http.MethodPost,
			userID:              1,
			contentType:         "text/plain",
			body:                "123456789",
			isInsert:            false,
			errInsert:           repository.ErrConflict,
			expectedContentType: "application/json",
			expectedStatus:      http.StatusConflict,
			expectedBody:        "{\"error\":\"order already uploaded by another user\"}\n",
		},
		{
			name:                "error_service",
			method:              http.MethodPost,
			userID:              1,
			contentType:         "text/plain",
			body:                "123456789",
			isInsert:            false,
			errInsert:           errors.New("internal server erro"),
			expectedContentType: "application/json",
			expectedStatus:      http.StatusInternalServerError,
			expectedBody:        "{\"error\":\"internal server error\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/", strings.NewReader(test.body))
			req.Header.Set("Content-Type", test.contentType)

			if test.userID != 0 {
				ctx := context.WithValue(t.Context(), models.UserIDContextKey, test.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			mos := mockhandler.NewMockOrderServer(t)
			mos.On("LoadUserOrder", req.Context(), mock.Anything, mock.Anything).Return(test.isInsert, test.errInsert).Maybe()
			h := NewOrderHandler(&cfg, mos)

			h.UploadUserOrder(w, req)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatus, w.Code)
			assert.Equal(t, test.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, test.expectedBody, string(body))
		})
	}
}
