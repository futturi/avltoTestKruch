package service

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/models"
	"testAlvtoShp/internal/store"
	"time"
)

type AuthService struct {
	store     store.Auth
	jwtSecret string
}

func NewAuthService(store store.Auth, jwtSecret string) *AuthService {
	return &AuthService{store: store, jwtSecret: jwtSecret}
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (int, error) {
	return s.store.GetUserByUsername(ctx, username)
}

func (s *AuthService) GenerateToken(id int) (string, error) {
	claims := jwt.MapClaims{
		"userId": id,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) CreateUser(ctx context.Context, req models.AuthRequest) (int, error) {
	req.HashPass()
	return s.store.CreateUser(ctx, req)
}

func (s *AuthService) CheckPassword(ctx context.Context, req models.AuthRequest) (bool, error) {
	req.HashPass()
	return s.store.CheckPassword(ctx, req)
}

func (s *AuthService) ExtractUserIDFromAccessToken(ctx context.Context, accessToken string) (int, error) {
	log := logger.LoggerFromContext(ctx)
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		log.Errorw("invalid access token", zap.Error(err))
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Errorw("invalid access token", zap.Error(err))
		return 0, jwt.ErrInvalidKey
	}

	userIDFloat, ok := claims["userId"].(float64)
	if !ok {
		log.Errorw("invalid access token, userId not found in token", zap.Error(err))
		return 0, errors.New("userId not found in token")
	}

	ok, err = s.store.CheckUserId(ctx, int(userIDFloat))

	if err != nil {
		log.Errorw("userId not found", zap.Error(err))
		return 0, err
	}

	if !ok {
		return 0, errors.New("userId not found in token")
	}

	return int(userIDFloat), nil
}
