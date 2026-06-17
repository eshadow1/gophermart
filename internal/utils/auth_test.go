package utils

import (
	"context"
	"testing"

	"github.com/eshadow1/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		expectUserID  int64
		expectIsValid bool
	}{
		{
			name:          "success",
			userID:        1,
			expectUserID:  1,
			expectIsValid: true,
		},
		{
			name:          "empty",
			userID:        0,
			expectUserID:  0,
			expectIsValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			if tc.userID != 0 {
				ctx = context.WithValue(t.Context(), models.UserIDContextKey, tc.userID)
			}

			userID, ok := GetUserID(ctx)
			assert.Equal(t, tc.expectIsValid, ok)
			assert.Equal(t, tc.expectUserID, userID)
		})
	}
}
