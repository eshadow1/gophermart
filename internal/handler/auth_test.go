package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/models"
	mockhandler "github.com/eshadow1/gophermart/mocks/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_RegisterUser(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name           string
		method         string
		token          string
		errLogin       error
		body           string
		expectedStatus int
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			token:          "",
			errLogin:       nil,
			expectedStatus: http.StatusOK,
			body: `{
						"login": "<login>",
						"password": "<password>"
					} `,
		},
		{
			name:           "bad_data",
			method:         http.MethodPost,
			token:          "",
			errLogin:       nil,
			expectedStatus: http.StatusBadRequest,
			body: `{
						"login": "<login>",
						"password": "<password>",
					} `,
		},
		{
			name:           "error_register",
			method:         http.MethodPost,
			token:          "",
			errLogin:       models.ErrInvalidFormat,
			expectedStatus: http.StatusBadRequest,
			body: `{
						"login": "<login>",
						"password": "<password>"
					} `,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/api/user/login", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			mas := mockhandler.NewMockAuthService(t)
			mas.On("Register", t.Context(), mock.Anything, mock.Anything).Return(test.token, test.errLogin).Maybe()
			h := NewAuthHandler(mas)

			h.RegisterUser(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_LoginUser(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name           string
		method         string
		token          string
		errLogin       error
		body           string
		expectedStatus int
	}{
		{
			name:           "success",
			method:         http.MethodPost,
			token:          "",
			errLogin:       nil,
			expectedStatus: http.StatusOK,
			body: `{
						"login": "<login>",
						"password": "<password>"
					} `,
		},
		{
			name:           "bad_data",
			method:         http.MethodPost,
			token:          "",
			errLogin:       nil,
			expectedStatus: http.StatusBadRequest,
			body: `{
						"login": "<login>",
						"unknown": "<password>",
					} `,
		},
		{
			name:           "error_login",
			method:         http.MethodPost,
			token:          "",
			errLogin:       models.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			body: `{
						"login": "<login>",
						"password": "<password>"
					} `,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), test.method, "/api/user/login", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			mas := mockhandler.NewMockAuthService(t)
			mas.On("Login", t.Context(), mock.Anything, mock.Anything).Return(test.token, test.errLogin).Maybe()
			h := NewAuthHandler(mas)

			h.LoginUser(w, req)

			assert.Equal(t, test.expectedStatus, w.Code)
		})
	}
}
