package handler

import (
	"testing"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/loggers"
	mockhandler "github.com/eshadow1/gophermart/mocks/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitRouter(t *testing.T) {
	cfg := configs.Config{}
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name string
	}{
		{
			name: "success_no_content",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			order := mockhandler.NewMockOrder(t)
			balance := mockhandler.NewMockBalance(t)
			auth := mockhandler.NewMockAuthHandler(t)

			chi := InitRouter(&cfg, order, balance, auth)
			assert.NotNil(t, chi)
		})
	}
}
