package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"
)

// @Description Возможные варианты покупки
type ItemForBuy string

const (
	ItemTShirt    ItemForBuy = "t-shirt"
	ItemCup       ItemForBuy = "cup"
	ItemBook      ItemForBuy = "book"
	ItemPen       ItemForBuy = "pen"
	ItemPowerbank ItemForBuy = "powerbank"
	ItemHoody     ItemForBuy = "hoody"
	ItemUmbrella  ItemForBuy = "umbrella"
	ItemSocks     ItemForBuy = "socks"
	ItemWallet    ItemForBuy = "wallet"
	ItemPinkHoody ItemForBuy = "pink-hoody"
)

// @Description Запрос на вход/регистрацию
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// @Description Ответ на вход/регистрацию
type AuthResponse struct {
	Token string `json:"token"`
}

// @Description Ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"errors"`
}

// @Description Информация о пользователе
type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Item      `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

// @Description Параметры айтема
type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

// @Description Запрос на перевод коинов
type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

// @Description История переводов коинов
type CoinHistory struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}

// @Description Полученные коины
type ReceivedTransaction struct {
	FromUser string `json:"fromUser" db:"from_user"`
	Amount   int    `json:"amount"`
}

// @Description Отправленные коины
type SentTransaction struct {
	ToUser string `json:"toUser"  db:"to_user"`
	Amount int    `json:"amount"`
}

type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Coins        int    `db:"coins"`
}

type Inventory struct {
	ID       int64  `db:"id"`
	UserID   int64  `db:"user_id"`
	ItemType string `db:"item_type"`
	Quantity int    `db:"quantity"`
}

type CoinTransaction struct {
	ID         int64         `db:"id"`
	SenderID   sql.NullInt64 `db:"sender_id"`
	ReceiverID sql.NullInt64 `db:"receiver_id"`
	Amount     int           `db:"amount"`
	Timestamp  time.Time     `db:"timestamp"`
}

func (a *AuthRequest) HashPass() {
	hash := sha256.Sum256([]byte(a.Password))
	a.Password = hex.EncodeToString(hash[:])
}
