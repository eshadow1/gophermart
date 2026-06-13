package repository

import (
	"errors"
	"testing"

	"github.com/eshadow1/gophermart/internal/loggers"
	mockrepository "github.com/eshadow1/gophermart/mocks/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestOrderRepo_AddOrder(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name               string
		userID             int64
		order              string
		errSelect          error
		errInsert          error
		expectedErrCreate  error
		expectedIsInserted bool
	}{
		{
			name:               "success",
			userID:             1,
			order:              "1",
			errSelect:          pgx.ErrNoRows,
			errInsert:          nil,
			expectedErrCreate:  nil,
			expectedIsInserted: true,
		},
		{
			name:               "error_select",
			userID:             1,
			order:              "1",
			errSelect:          errors.New("error_select"),
			errInsert:          nil,
			expectedErrCreate:  errors.New("error_select"),
			expectedIsInserted: false,
		},
		{
			name:               "error_insert",
			userID:             1,
			order:              "1",
			errSelect:          pgx.ErrNoRows,
			errInsert:          errors.New("error_insert"),
			expectedErrCreate:  errors.New("error_insert"),
			expectedIsInserted: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testRow := &MockRow{}
			testRow.On("Scan", mock.Anything, mock.Anything).Return(test.errSelect).Maybe()
			mockOrderDB := mockrepository.NewMockDBPool(t)
			mockOrderDB.On("Exec", t.Context(), mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, test.errInsert).Maybe()
			mockOrderDB.On("QueryRow", t.Context(), mock.Anything, mock.Anything).Return(testRow).Maybe()

			repo := NewOrderRepo(mockOrderDB)
			isInsert, errCreate := repo.AddOrder(t.Context(), test.userID, test.order)

			if test.expectedErrCreate != nil {
				assert.Equal(t, test.expectedErrCreate, errCreate)
			} else {
				require.NoError(t, errCreate)
			}
			assert.Equal(t, test.expectedIsInserted, isInsert)
		})
	}
}

func TestOrderRepo_UpdateOrderStatus(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name              string
		number            string
		status            string
		accual            float64
		errUpdate         error
		expectedErrUpdate error
	}{
		{
			name:              "success",
			number:            "1",
			status:            "new",
			accual:            1.0,
			errUpdate:         nil,
			expectedErrUpdate: nil,
		},
		{
			name:              "invalid_update",
			number:            "1",
			status:            "new",
			accual:            1.0,
			errUpdate:         errors.New("invalid_update"),
			expectedErrUpdate: errors.New("invalid_update"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOrderDB := mockrepository.NewMockDBPool(t)
			mockOrderDB.On("Exec", t.Context(), mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, test.errUpdate).Maybe()

			repo := NewOrderRepo(mockOrderDB)
			errUpdate := repo.UpdateOrderStatus(t.Context(), test.number, test.status, &test.accual)

			if test.expectedErrUpdate != nil {
				assert.Equal(t, test.expectedErrUpdate, errUpdate)
			} else {
				assert.NoError(t, errUpdate)
			}
		})
	}
}
