package service

import (
	"context"
	"testAlvtoShp/internal/models"
	"testAlvtoShp/internal/store"
)

type ShopService struct {
	store store.Shop
}

func NewShopService(store store.Shop) *ShopService {
	return &ShopService{store}
}

func (s *ShopService) GetUserInfo(ctx context.Context, userId int) (models.InfoResponse, error) {
	return s.store.GetUserInfo(ctx, userId)
}

func (s *ShopService) BuyItem(ctx context.Context, userId int, item string) error {
	return s.store.BuyItem(ctx, userId, item)
}

func (s *ShopService) SendCoin(ctx context.Context, userId int, req models.SendCoinRequest) error {
	return s.store.SendCoin(ctx, userId, req)
}
