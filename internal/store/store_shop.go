package store

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	internalErrors "testAlvtoShp/internal/errors"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/models"
)

type ShopStore struct {
	Db *sqlx.DB
}

func NewShopStore(db *sqlx.DB) *ShopStore {
	return &ShopStore{
		Db: db,
	}
}

func (r *ShopStore) GetUserInfo(ctx context.Context, userId int) (models.InfoResponse, error) {
	log := logger.LoggerFromContext(ctx)
	var user models.User
	if err := r.Db.Get(&user, fmt.Sprintf("SELECT id, username, coins FROM %s WHERE id=$1", usersTable), userId); err != nil {
		log.Errorw("GetUserInfo get coins", zap.Error(err))
		return models.InfoResponse{}, err
	}

	var inventories []models.Inventory
	if err := r.Db.Select(&inventories, fmt.Sprintf("SELECT id, user_id, item_type, quantity FROM %s WHERE user_id=$1", inventoryTable), userId); err != nil {
		log.Errorw("GetUserInfo get inventory", zap.Error(err))
		return models.InfoResponse{}, err
	}
	var inventoryItems []models.Item
	for _, inv := range inventories {
		inventoryItems = append(inventoryItems, models.Item{
			Type:     inv.ItemType,
			Quantity: inv.Quantity,
		})
	}

	var receivedTx []models.ReceivedTransaction
	queryReceived := fmt.Sprintf(`
			SELECT ct.amount, COALESCE(u.username, 'Unknown') AS from_user
			FROM %s ct
			LEFT JOIN users u ON ct.sender_id = u.id
			WHERE ct.receiver_id = $1
		`, coinTxTable)
	if err := r.Db.Select(&receivedTx, queryReceived, userId); err != nil {
		log.Errorw("GetUserInfo get story gotten tx", zap.Error(err))
		return models.InfoResponse{}, err
	}
	//todo юзера чекать в бд
	var received []models.ReceivedTransaction
	for _, rt := range receivedTx {
		received = append(received, models.ReceivedTransaction{
			FromUser: rt.FromUser,
			Amount:   rt.Amount,
		})
	}

	var sentTx []models.SentTransaction
	querySent := fmt.Sprintf(`
			SELECT ct.amount, COALESCE(u.username, 'Unknown') AS to_user
			FROM %s ct
			LEFT JOIN users u ON ct.receiver_id = u.id
			WHERE ct.sender_id = $1
		`, coinTxTable)
	if err := r.Db.Select(&sentTx, querySent, userId); err != nil {
		log.Errorw("GetUserInfo get story send tx", zap.Error(err))
		return models.InfoResponse{}, err
	}
	var sent []models.SentTransaction
	for _, st := range sentTx {
		sent = append(sent, models.SentTransaction{
			ToUser: st.ToUser,
			Amount: st.Amount,
		})
	}

	response := models.InfoResponse{
		Coins:     user.Coins,
		Inventory: inventoryItems,
		CoinHistory: models.CoinHistory{
			Received: received,
			Sent:     sent,
		},
	}

	return response, nil

}

func (r *ShopStore) BuyItem(ctx context.Context, userId int, item string) error {
	log := logger.LoggerFromContext(ctx)
	tx, err := r.Db.Begin()
	if err != nil {
		log.Errorw("failed to begin transaction", zap.Error(err))
		return err
	}

	firstQuery := fmt.Sprintf(`
	select coins from %s where id = $1
`, usersTable)

	var coins int

	err = tx.QueryRow(firstQuery, userId).Scan(&coins)

	if err != nil {
		log.Errorw("failed to scan row", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	secondQuery := fmt.Sprintf(`
	select price from %s where name = $1
`, itemsTable)

	var price int
	err = tx.QueryRow(secondQuery, item).Scan(&price)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		log.Errorw("failed to scan row", zap.Error(err))
		return err
	}

	if coins < price {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		log.Errorw("user doesnt have enough money")
		return internalErrors.NoMoney
	}

	thirdQuery := fmt.Sprintf(`
	update %s set coins = coins - $1 where id = $2
`, usersTable)

	_, err = tx.Exec(thirdQuery, price, userId)
	if err != nil {
		log.Errorw("failed to scan row", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	fourthQuery := fmt.Sprintf(`
	select count(*) from %s where user_id = $1 and item_type=$2
`, inventoryTable)

	row := tx.QueryRow(fourthQuery, userId, item)

	var countInventory int

	if err = row.Scan(&countInventory); err != nil {
		log.Errorw("failed to scan row", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	var fifthQuery string

	if countInventory == 0 {
		fifthQuery = fmt.Sprintf(`
		insert into %s (item_type, user_id, quantity) values($1,$2,1)
`, inventoryTable)

	} else {
		fifthQuery = fmt.Sprintf(`
	update %s set item_type = $1, quantity = quantity + 1 where user_id = $2
`, inventoryTable)
	}

	_, err = tx.Exec(fifthQuery, item, userId)

	if err != nil {
		log.Errorw("failed to scan row", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	return tx.Commit()

}

func (r *ShopStore) SendCoin(ctx context.Context, userId int, req models.SendCoinRequest) error {
	log := logger.LoggerFromContext(ctx)
	tx, err := r.Db.Begin()
	if err != nil {
		log.Errorw("failed to begin transaction", zap.Error(err))
		return err
	}

	fQuery := fmt.Sprintf(`
	select id from %s where username = $1
`, usersTable)

	var toUserId, coins int

	if err = tx.QueryRow(fQuery, req.ToUser).Scan(&toUserId); err != nil {
		log.Errorw("failed to scan row", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	sQuery := fmt.Sprintf(`
	select coins from %s where id = $1
`, usersTable)

	if err = tx.QueryRow(sQuery, userId).Scan(&coins); err != nil {
		log.Errorw("failed to scan row", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}

		return err
	}

	if coins < req.Amount {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		log.Errorw("user doesnt have enough money", zap.Int("amount", req.Amount), zap.Int("userCoins", coins))
		return internalErrors.NoMoney
	}

	firstQuery := fmt.Sprintf(`
	update %s set coins = coins + $1 where id = $2
`, usersTable)

	_, err = tx.Exec(firstQuery, req.Amount, toUserId)

	if err != nil {
		log.Errorw("failed to send coin", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	secondQuery := fmt.Sprintf(`
	update %s set coins = coins - $1 where id = $2
`, usersTable)

	_, err = tx.Exec(secondQuery, req.Amount, userId)

	if err != nil {
		log.Errorw("failed to send coin", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	thirdQuery := fmt.Sprintf(`
	insert into %s(sender_id, receiver_id, amount) values($1, $2, $3)
`, coinTxTable)

	_, err = tx.Exec(thirdQuery, userId, toUserId, req.Amount)
	if err != nil {
		log.Errorw("failed to insert coin", zap.Error(err))
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Errorw("failed to rollback transaction", zap.Error(rollbackErr))
		}
		return err
	}

	return tx.Commit()

}
