package configs

import (
	"testing"

	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Init(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name                  string
		addrSystem            string
		addr                  string
		logLevel              string
		storagePathDB         string
		storagePathMigrations string
		authJWTSecret         []byte
		authTokenIssuer       string
	}{
		{
			name:                  "success",
			addrSystem:            DefaultSystemAddr,
			addr:                  DefaultAddr,
			logLevel:              DefaultLevelLog,
			storagePathDB:         DefaultEmptyString,
			storagePathMigrations: DefaultMigrationPath,
			authJWTSecret:         []byte(DefaultEmptyString),
			authTokenIssuer:       DefaultEmptyString,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.Init()

			assert.Equal(t, test.addrSystem, cfg.AddrSystem)
			assert.Equal(t, test.addr, cfg.Addr)
			assert.Equal(t, test.logLevel, cfg.Log.Level)
			assert.Equal(t, test.storagePathDB, cfg.Storage.PathDB)
			assert.Equal(t, test.storagePathMigrations, cfg.Storage.PathMigrations)
			assert.Equal(t, test.authJWTSecret, cfg.Auth.JWTSecret)
			assert.Equal(t, test.authTokenIssuer, cfg.Auth.TokenIssuer)
		})
	}
}
