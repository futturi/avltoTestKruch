package store

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"errors"
	"testAlvtoShp/internal/models"
)

func TestShopStore_GetUserInfo(t *testing.T) {
	tests := []struct {
		name           string
		userId         int
		setupMock      func(mock sqlmock.Sqlmock)
		expectedResult models.InfoResponse
		expectErr      bool
	}{
		{
			name:   "Error in user query",
			userId: 42,
			setupMock: func(mock sqlmock.Sqlmock) {
				queryUser := regexp.MustCompile(`(?i)^SELECT id,\s*username,\s*coins FROM\s+` + usersTable + `\s+WHERE id=\$1$`)
				mock.ExpectQuery(queryUser.String()).WithArgs(42).
					WillReturnError(errors.New("db error"))
			},
			expectedResult: models.InfoResponse{},
			expectErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			shopStore := NewShopStore(sqlxDB)

			tc.setupMock(mock)

			result, err := shopStore.GetUserInfo(context.Background(), tc.userId)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestShopStore_BuyItem(t *testing.T) {
	tests := []struct {
		name           string
		userId         int
		item           string
		coins          int
		price          int
		invCount       int
		setupMock      func(mock sqlmock.Sqlmock, userId int, item string, coins, price, invCount int)
		expectErr      bool
		expectedErrVal error
	}{
		{
			name:     "Success_Insert (нет записи в инвентаре)",
			userId:   42,
			item:     "sword",
			coins:    100,
			price:    50,
			invCount: 0,
			setupMock: func(mock sqlmock.Sqlmock, userId int, item string, coins, price, invCount int) {
				mock.ExpectBegin()

				queryCoins := regexp.MustCompile(`(?i)^select coins from\s+` + usersTable + `\s+where id = \$1$`)
				mock.ExpectQuery(queryCoins.String()).WithArgs(userId).
					WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(coins))

				queryPrice := regexp.MustCompile(`(?i)^select price from\s+` + itemsTable + `\s+where name = \$1$`)
				mock.ExpectQuery(queryPrice.String()).WithArgs(item).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(price))

				queryUpdateCoins := regexp.MustCompile(`(?i)^update\s+` + usersTable + `\s+set coins = coins - \$1 where id = \$2$`)
				mock.ExpectExec(queryUpdateCoins.String()).WithArgs(price, userId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				queryInvCount := regexp.MustCompile(`(?i)^select count\(\*\) from\s+` + inventoryTable + `\s+where user_id = \$1 and item_type=\$2$`)
				mock.ExpectQuery(queryInvCount.String()).WithArgs(userId, item).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(invCount))

				if invCount == 0 {
					_ = regexp.MustCompile(`(?i)^insert into\s+` + inventoryTable + `\s+\(item_type, user_id, quantity\)\s+values\(\$1,\$2,1\)`).
						String()
				} else {
					_ = regexp.MustCompile(`(?i)^update\s+` + inventoryTable + `\s+set item_type = \$1, quantity = quantity \+ 1 where user_id = \$2`).
						String()
				}
				queryInvRegex := regexp.MustCompile(`(?i)^(insert into|update)\s+` + inventoryTable)
				mock.ExpectExec(queryInvRegex.String()).WithArgs(item, userId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
			expectErr: false,
		},
		{
			name:     "Success_Update (есть запись в инвентаре)",
			userId:   42,
			item:     "shield",
			coins:    100,
			price:    50,
			invCount: 1,
			setupMock: func(mock sqlmock.Sqlmock, userId int, item string, coins, price, invCount int) {
				mock.ExpectBegin()

				queryCoins := regexp.MustCompile(`(?i)^select coins from\s+` + usersTable + `\s+where id = \$1$`)
				mock.ExpectQuery(queryCoins.String()).WithArgs(userId).
					WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(coins))

				queryPrice := regexp.MustCompile(`(?i)^select price from\s+` + itemsTable + `\s+where name = \$1$`)
				mock.ExpectQuery(queryPrice.String()).WithArgs(item).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(price))

				queryUpdateCoins := regexp.MustCompile(`(?i)^update\s+` + usersTable + `\s+set coins = coins - \$1 where id = \$2$`)
				mock.ExpectExec(queryUpdateCoins.String()).WithArgs(price, userId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				queryInvCount := regexp.MustCompile(`(?i)^select count\(\*\) from\s+` + inventoryTable + `\s+where user_id = \$1 and item_type=\$2$`)
				mock.ExpectQuery(queryInvCount.String()).WithArgs(userId, item).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(invCount))

				queryInvRegex := regexp.MustCompile(`(?i)^update\s+` + inventoryTable)
				mock.ExpectExec(queryInvRegex.String()).WithArgs(item, userId).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
			expectErr: false,
		},
		{
			name:     "Insufficient funds",
			userId:   42,
			item:     "potion",
			coins:    40,
			price:    50,
			invCount: 0,
			setupMock: func(mock sqlmock.Sqlmock, userId int, item string, coins, price, invCount int) {
				mock.ExpectBegin()

				queryCoins := regexp.MustCompile(`(?i)^select coins from\s+` + usersTable + `\s+where id = \$1$`)
				mock.ExpectQuery(queryCoins.String()).WithArgs(userId).
					WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(coins))

				queryPrice := regexp.MustCompile(`(?i)^select price from\s+` + itemsTable + `\s+where name = \$1$`)
				mock.ExpectQuery(queryPrice.String()).WithArgs(item).
					WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(price))

				mock.ExpectRollback()
			},
			expectErr:      true,
			expectedErrVal: errors.New(""),
		},
		{
			name:     "Error in second query (price)",
			userId:   42,
			item:     "axe",
			coins:    100,
			price:    0,
			invCount: 0,
			setupMock: func(mock sqlmock.Sqlmock, userId int, item string, coins, price, invCount int) {
				mock.ExpectBegin()

				queryCoins := regexp.MustCompile(`(?i)^select coins from\s+` + usersTable + `\s+where id = \$1$`)
				mock.ExpectQuery(queryCoins.String()).WithArgs(userId).
					WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(coins))

				queryPrice := regexp.MustCompile(`(?i)^select price from\s+` + itemsTable + `\s+where name = \$1$`)
				mock.ExpectQuery(queryPrice.String()).WithArgs(item).
					WillReturnError(errors.New("price query error"))

				mock.ExpectRollback()
			},
			expectErr:      true,
			expectedErrVal: errors.New(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			shopStore := NewShopStore(sqlxDB)

			tc.setupMock(mock, tc.userId, tc.item, tc.coins, tc.price, tc.invCount)

			err = shopStore.BuyItem(context.Background(), tc.userId, tc.item)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestShopStore_SendCoin(t *testing.T) {
	tests := []struct {
		name           string
		userId         int
		req            models.SendCoinRequest
		toUserId       int
		toUserCoins    int
		setupMock      func(mock sqlmock.Sqlmock, userId int, req models.SendCoinRequest, toUserId, toUserCoins int)
		expectErr      bool
		expectedErrVal error
	}{
		{
			name:   "Insufficient funds for receiver",
			userId: 42,
			req: models.SendCoinRequest{
				ToUser: "alice",
				Amount: 50,
			},
			toUserId:    100,
			toUserCoins: 30,
			setupMock: func(mock sqlmock.Sqlmock, userId int, req models.SendCoinRequest, toUserId, toUserCoins int) {
				mock.ExpectBegin()

				queryF := regexp.MustCompile(`(?i)^select id from\s+` + usersTable + `\s+where username = \$1$`)
				mock.ExpectQuery(queryF.String()).WithArgs(req.ToUser).
					WillReturnRows(sqlmock.NewRows([]string{"id", "coins"}).AddRow(toUserId, toUserCoins))
				mock.ExpectRollback()
			},
			expectErr:      true,
			expectedErrVal: errors.New(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")
			shopStore := NewShopStore(sqlxDB)

			tc.setupMock(mock, tc.userId, tc.req, tc.toUserId, tc.toUserCoins)

			err = shopStore.SendCoin(context.Background(), tc.userId, tc.req)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
