package utils

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondError(t *testing.T) {
	tests := []struct {
		name                string
		err                 error
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:                "validation_error",
			err:                 &ValidationError{Msg: "validation error"},
			expectedStatus:      http.StatusBadRequest,
			expectedContentType: "application/json",
			expectedBody:        "{\"error\":\"validation error\"}\n",
		},
		{
			name:                "incorrect_format_error",
			err:                 &IncorrectFormatError{Msg: "incorrect format error"},
			expectedStatus:      http.StatusUnprocessableEntity,
			expectedContentType: "application/json",
			expectedBody:        "{\"error\":\"incorrect format error\"}\n",
		},
		{
			name:                "conflict_error",
			err:                 &ConflictError{Msg: "conflict error"},
			expectedStatus:      http.StatusConflict,
			expectedContentType: "application/json",
			expectedBody:        "{\"error\":\"conflict error\"}\n",
		},
		{
			name:                "unauthorized_error",
			err:                 &UnauthorizedError{Msg: "unauthorized error"},
			expectedStatus:      http.StatusUnauthorized,
			expectedContentType: "application/json",
			expectedBody:        "{\"error\":\"unauthorized error\"}\n",
		},
		{
			name:                "insufficient_funds_error",
			err:                 &InsufficientFundsError{Msg: "insufficient funds error"},
			expectedStatus:      http.StatusPaymentRequired,
			expectedContentType: "application/json",
			expectedBody:        "{\"error\":\"insufficient funds error\"}\n",
		},
		{
			name:                "internal_server_error",
			err:                 errors.New("internal server error"),
			expectedStatus:      http.StatusInternalServerError,
			expectedContentType: "application/json",
			expectedBody:        "{\"error\":\"internal server error\"}\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			RespondError(w, tc.err)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedContentType, w.Header().Get("Content-Type"))
			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
