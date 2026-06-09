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

type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}

func TestUserRepo_Create(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name              string
		user              *models.User
		errInsert         error
		expectedErrCreate error
	}{
		{
			name: "success",
			user: &models.User{
				Login:        "user",
				ID:           1,
				PasswordHash: "user",
			},
			errInsert:         nil,
			expectedErrCreate: nil,
		},
		{
			name: "error_insert",
			user: &models.User{
				Login:        "user",
				ID:           1,
				PasswordHash: "user",
			},
			errInsert: &pgconn.PgError{
				Code: codePostgresDuplicateInsert,
			},
			expectedErrCreate: models.ErrUserAlreadyExists,
		},
		{
			name: "error_insert",
			user: &models.User{
				Login:        "user",
				ID:           1,
				PasswordHash: "user",
			},
			errInsert:         errors.New("error"),
			expectedErrCreate: errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockUserDB := mockrepository.NewMockUserDBPool(t)
			mockUserDB.On("Exec", t.Context(), mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, test.errInsert).Maybe()
			repo := NewUserRepo(mockUserDB)

			errCreate := repo.Create(t.Context(), test.user)

			if test.expectedErrCreate != nil {
				assert.Equal(t, test.expectedErrCreate, errCreate)
			} else {
				assert.NoError(t, errCreate)
			}
		})
	}
}

func TestUserRepo_GetByLogin(t *testing.T) {
	errLog := loggers.CreateLogger("Debug")
	require.NoError(t, errLog)

	tests := []struct {
		name              string
		login             string
		errSelect         error
		expectedErrCreate error
	}{
		{
			name:              "success",
			login:             "user",
			errSelect:         nil,
			expectedErrCreate: nil,
		},
		{
			name:              "invalid_credentials",
			login:             "user",
			errSelect:         pgx.ErrNoRows,
			expectedErrCreate: models.ErrInvalidCredentials,
		},
		{
			name:              "error_select",
			login:             "user",
			errSelect:         errors.New("error"),
			expectedErrCreate: errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testRow := &MockRow{}
			testRow.On("Scan", mock.Anything, mock.Anything).Return(test.errSelect).Maybe()
			mockUserDB := mockrepository.NewMockUserDBPool(t)
			mockUserDB.On("QueryRow", t.Context(), mock.Anything, mock.Anything).Return(testRow).Maybe()

			repo := NewUserRepo(mockUserDB)

			_, errCreate := repo.GetByLogin(t.Context(), test.login)

			if test.expectedErrCreate != nil {
				assert.Equal(t, test.expectedErrCreate, errCreate)
			} else {
				assert.NoError(t, errCreate)
			}
		})
	}
}
