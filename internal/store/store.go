package store

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"testAlvtoShp/internal/config"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/models"
)

const (
	usersTable     = "users"
	inventoryTable = "inventory"
	coinTxTable    = "coin_transactions"
	itemsTable     = "items"
)

func NewDbConn(ctx context.Context, cfg *config.Config) (*sqlx.DB, error) {
	log := logger.LoggerFromContext(ctx)
	log.Debugw("connecting to database")
	conn, err := sqlx.Connect("postgres", fmt.Sprintf("host =%s port =%s user =%s dbname=%s password=%s sslmode=%s",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbName, cfg.DbPass, "disable"))
	if err != nil {
		log.Errorw("failed to connect to database", zap.Error(err))
		return nil, err
	}

	return conn, nil
}

func ShutDown(ctx context.Context, db *sqlx.DB) error {
	log := logger.LoggerFromContext(ctx)
	log.Debugw("shutting down database")
	err := db.Close()
	if err != nil {
		log.Errorw("failed to close database", zap.Error(err))
		return err
	}

	return nil
}

type Store struct {
	Auth
	Shop
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{
		Auth: NewAuthStore(db),
		Shop: NewShopStore(db),
	}
}

type Auth interface {
	GetUserByUsername(ctx context.Context, username string) (int, error)
	CreateUser(ctx context.Context, req models.AuthRequest) (int, error)
	CheckPassword(ctx context.Context, req models.AuthRequest) (bool, error)
	CheckUserId(ctx context.Context, userId int) (bool, error)
}

type Shop interface {
	GetUserInfo(ctx context.Context, userId int) (models.InfoResponse, error)
	BuyItem(ctx context.Context, userId int, item string) error
	SendCoin(ctx context.Context, userId int, req models.SendCoinRequest) error
}
