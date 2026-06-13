package repository

import (
	"errors"
	"testing"

	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/eshadow1/gophermart/internal/models"
	mockrepository "github.com/eshadow1/gophermart/mocks/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBalanceRepo_Get(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name            string
		userID          int64
		errExec         error
		expectedErrGet  error
		expectedBalance models.BalanceResponse
	}{
		{
			name:            "success empty",
			userID:          1,
			errExec:         pgx.ErrNoRows,
			expectedErrGet:  nil,
			expectedBalance: models.BalanceResponse{},
		},
		{
			name:            "success",
			userID:          1,
			errExec:         nil,
			expectedErrGet:  nil,
			expectedBalance: models.BalanceResponse{},
		},
		{
			name:            "error_insert",
			userID:          1,
			errExec:         errors.New("error"),
			expectedErrGet:  errors.New("error"),
			expectedBalance: models.BalanceResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testRow := &MockRow{}
			testRow.On("Scan", mock.Anything, mock.Anything).Return(test.errExec).Maybe()

			mockBalanceDB := mockrepository.NewMockDBPool(t)
			mockBalanceDB.On("QueryRow", t.Context(), mock.Anything, mock.Anything).Return(testRow).Maybe()
			repo := NewBalanceRepo(mockBalanceDB)
			balance, errCreate := repo.Get(t.Context(), test.userID)

			if test.expectedErrGet != nil {
				assert.Equal(t, test.expectedErrGet, errCreate)
			} else {
				require.NoError(t, errCreate)
			}
			assert.Equal(t, test.expectedBalance, balance)
		})
	}
}

func TestBalanceRepo_AddAccrual(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name               string
		userID             int64
		amount             float64
		errExec            error
		expectedErrAccrual error
	}{
		{
			name:               "success_insert",
			userID:             1,
			amount:             0,
			errExec:            nil,
			expectedErrAccrual: nil,
		},
		{
			name:               "error_insert",
			userID:             1,
			amount:             0,
			errExec:            errors.New("error"),
			expectedErrAccrual: errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockBalanceDB := mockrepository.NewMockDBPool(t)
			mockBalanceDB.On("Exec", t.Context(), mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, test.errExec).Maybe()
			repo := NewBalanceRepo(mockBalanceDB)

			errAccrual := repo.AddAccrual(t.Context(), test.userID, test.amount)

			if test.expectedErrAccrual != nil {
				assert.Equal(t, test.expectedErrAccrual, errAccrual)
			} else {
				assert.NoError(t, errAccrual)
			}
		})
	}
}
