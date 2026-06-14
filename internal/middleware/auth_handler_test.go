package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	cfg := &configs.AuthConfig{
		JWTSecret:   []byte(configs.DefaultEmptyString),
		TokenIssuer: configs.DefaultEmptyString,
	}
	userID := int64(1)
	token, errGenerate := generateTestToken(cfg.JWTSecret, userID)
	require.NoError(t, errGenerate)

	tests := []struct {
		name         string
		token        string
		isValid      bool
		cookieToken  string
		expectStatus int
	}{
		{
			name:         "success_token",
			token:        "Bearer ",
			isValid:      true,
			cookieToken:  "",
			expectStatus: http.StatusOK,
		},
		{
			name:         "success_cookie",
			token:        "",
			isValid:      true,
			cookieToken:  "auth_token=",
			expectStatus: http.StatusOK,
		},
		{
			name:         "error_parse",
			token:        "Bearer ",
			isValid:      false,
			cookieToken:  "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "error_extract",
			token:        "",
			cookieToken:  "",
			isValid:      true,
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			h := AuthMiddleware(cfg)(next)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
			if tc.token != "" {
				if tc.isValid {
					req.Header.Set("Authorization", tc.token+token)
				} else {
					req.Header.Set("Authorization", tc.token+token+"in")
				}
			}
			if tc.cookieToken != "" {
				req.Header.Set("Cookie", tc.cookieToken+token)
			}

			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
		})
	}
}

func generateTestToken(secretKey []byte, userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
