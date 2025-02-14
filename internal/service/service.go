package service

import (
	"context"
	"testAlvtoShp/internal/config"
	"testAlvtoShp/internal/models"
	"testAlvtoShp/internal/store"
)

type Service struct {
	Auth
	Shop
}

func NewService(store *store.Store, cfg *config.Config) *Service {
	return &Service{
		Auth: NewAuthService(store.Auth, cfg.JwtSecret),
		Shop: NewShopService(store.Shop),
	}
}

type Auth interface {
	GetUserByUsername(ctx context.Context, username string) (int, error)
	GenerateToken(id int) (string, error)
	CreateUser(ctx context.Context, req models.AuthRequest) (int, error)
	CheckPassword(ctx context.Context, req models.AuthRequest) (bool, error)
	ExtractUserIDFromAccessToken(ctx context.Context, accessToken string) (int, error)
}

type Shop interface {
	GetUserInfo(ctx context.Context, userId int) (models.InfoResponse, error)
	BuyItem(ctx context.Context, userId int, item string) error
	SendCoin(ctx context.Context, userId int, req models.SendCoinRequest) error
}
