package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLuhn(t *testing.T) {
	tests := []struct {
		name          string
		number        string
		expectIsValid bool
	}{
		{
			name:          "success_valid_luhn",
			number:        "6141023187502264",
			expectIsValid: true,
		},
		{
			name:          "empty_data",
			number:        "",
			expectIsValid: false,
		},
		{
			name:          "incorrect_luhn_data",
			number:        "unknown",
			expectIsValid: false,
		},
		{
			name:          "incorrect_luhn_number",
			number:        "614102318750226",
			expectIsValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isValid := ValidateLuhn(tc.number)
			assert.Equal(t, tc.expectIsValid, isValid)
		})
	}
}
