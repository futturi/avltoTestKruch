package store

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"testAlvtoShp/internal/models"
)

func TestAuthStore_GetUserByUsername(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		setupMock  func(mock sqlmock.Sqlmock)
		expectedID int
		expectErr  bool
	}{
		{
			name:     "Success",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select id from\s+` + usersTable + `\s+where username = \$1.*$`)
				rows := sqlmock.NewRows([]string{"id"}).AddRow(42)
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser").WillReturnRows(rows)
			},
			expectedID: 42,
			expectErr:  false,
		},
		{
			name:     "No rows found",
			username: "unknown",
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select id from\s+` + usersTable + `\s+where username = \$1.*$`)
				rows := sqlmock.NewRows([]string{"id"})
				mock.ExpectQuery(queryRegex.String()).WithArgs("unknown").WillReturnRows(rows)
			},
			expectedID: 0,
			expectErr:  false,
		},
		{
			name:     "DB error",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select id from\s+` + usersTable + `\s+where username = \$1.*$`)
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser").WillReturnError(errors.New("db error"))
			},
			expectedID: 0,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			authStore := NewAuthStore(sqlxDB)

			tc.setupMock(mock)

			id, err := authStore.GetUserByUsername(context.Background(), tc.username)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthStore_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		req        models.AuthRequest
		setupMock  func(mock sqlmock.Sqlmock)
		expectedID int
		expectErr  bool
	}{
		{
			name: "Success",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "hashedpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*insert into\s+` + usersTable + `\s+\(username, password_hash\)\s+values\s+\(\$1, \$2\)\s+returning id;.*$`)
				rows := sqlmock.NewRows([]string{"id"}).AddRow(42)
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser", "hashedpass").WillReturnRows(rows)
			},
			expectedID: 42,
			expectErr:  false,
		},
		{
			name: "Insert error",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "hashedpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*insert into\s+` + usersTable + `\s+\(username, password_hash\)\s+values\s+\(\$1, \$2\)\s+returning id;.*$`)
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser", "hashedpass").WillReturnError(errors.New("insert error"))
			},
			expectedID: 0,
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			authStore := NewAuthStore(sqlxDB)

			tc.setupMock(mock)

			id, err := authStore.CreateUser(context.Background(), tc.req)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthStore_CheckPassword(t *testing.T) {
	tests := []struct {
		name          string
		req           models.AuthRequest
		setupMock     func(mock sqlmock.Sqlmock)
		expectedMatch bool
		expectErr     bool
	}{
		{
			name: "Password match",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "hashedpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select password_hash from\s+` + usersTable + `\s+where username = \$1.*$`)
				rows := sqlmock.NewRows([]string{"password_hash"}).AddRow("hashedpass")
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser").WillReturnRows(rows)
			},
			expectedMatch: true,
			expectErr:     false,
		},
		{
			name: "Password mismatch",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "hashedpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select password_hash from\s+` + usersTable + `\s+where username = \$1.*$`)
				rows := sqlmock.NewRows([]string{"password_hash"}).AddRow("differentpass")
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser").WillReturnRows(rows)
			},
			expectedMatch: false,
			expectErr:     false,
		},
		{
			name: "No rows found",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "hashedpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select password_hash from\s+` + usersTable + `\s+where username = \$1.*$`)
				rows := sqlmock.NewRows([]string{"password_hash"})
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser").WillReturnRows(rows)
			},
			expectedMatch: false,
			expectErr:     false,
		},
		{
			name: "Query error",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "hashedpass",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select password_hash from\s+` + usersTable + `\s+where username = \$1.*$`)
				mock.ExpectQuery(queryRegex.String()).WithArgs("testuser").WillReturnError(errors.New("db error"))
			},
			expectedMatch: false,
			expectErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			authStore := NewAuthStore(sqlxDB)

			tc.setupMock(mock)

			match, err := authStore.CheckPassword(context.Background(), tc.req)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedMatch, match)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAuthStore_CheckUserId(t *testing.T) {
	tests := []struct {
		name         string
		userId       int
		setupMock    func(mock sqlmock.Sqlmock)
		expectedBool bool
		expectErr    bool
	}{
		{
			name:   "User exists",
			userId: 42,
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select count\(\*\) from\s+` + usersTable + `\s+where id = \$1.*$`)
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(queryRegex.String()).WithArgs(42).WillReturnRows(rows)
			},
			expectedBool: true,
			expectErr:    false,
		},
		{
			name:   "User does not exist",
			userId: 42,
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select count\(\*\) from\s+` + usersTable + `\s+where id = \$1.*$`)
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(queryRegex.String()).WithArgs(42).WillReturnRows(rows)
			},
			expectedBool: false,
			expectErr:    false,
		},
		{
			name:   "Query error",
			userId: 42,
			setupMock: func(mock sqlmock.Sqlmock) {
				queryRegex := regexp.MustCompile(`(?i)^.*select count\(\*\) from\s+` + usersTable + `\s+where id = \$1.*$`)
				mock.ExpectQuery(queryRegex.String()).WithArgs(42).WillReturnError(errors.New("db error"))
			},
			expectedBool: false,
			expectErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			authStore := NewAuthStore(sqlxDB)

			tc.setupMock(mock)

			ok, err := authStore.CheckUserId(context.Background(), tc.userId)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedBool, ok)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
