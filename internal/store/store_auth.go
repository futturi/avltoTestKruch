package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/models"
)

type AuthStore struct {
	Db *sqlx.DB
}

func NewAuthStore(db *sqlx.DB) *AuthStore {
	return &AuthStore{
		Db: db,
	}
}

func (r *AuthStore) GetUserByUsername(ctx context.Context, username string) (int, error) {
	log := logger.LoggerFromContext(ctx)
	query := fmt.Sprintf(`
	select id from %s where username = $1
`, usersTable)
	var id int

	row := r.Db.QueryRow(query, username)

	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		log.Errorw("error with scanning row", zap.Error(err))
		return 0, err
	}

	return id, nil
}

func (r *AuthStore) CreateUser(ctx context.Context, req models.AuthRequest) (int, error) {
	log := logger.LoggerFromContext(ctx)
	query := fmt.Sprintf(`

insert into %s (username, password_hash) values ($1, $2) returning id;`, usersTable)

	var id int

	err := r.Db.QueryRow(query, req.Username, req.Password).Scan(&id)

	if err != nil {
		log.Errorw("error with inserting row", zap.Error(err))
		return 0, err
	}

	return id, nil
}

func (r *AuthStore) CheckPassword(ctx context.Context, req models.AuthRequest) (bool, error) {
	log := logger.LoggerFromContext(ctx)
	query := fmt.Sprintf(`

select password_hash from %s where username = $1
`, usersTable)

	var password string

	row := r.Db.QueryRow(query, req.Username)

	if err := row.Scan(&password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		log.Errorw("error with scanning row", zap.Error(err))
		return false, err
	}

	if password != req.Password {
		return false, nil
	}

	return true, nil
}

func (r *AuthStore) CheckUserId(ctx context.Context, userId int) (bool, error) {
	log := logger.LoggerFromContext(ctx)

	query := fmt.Sprintf(`
	select count(*) from %s where id = $1
`, usersTable)

	var count int

	row := r.Db.QueryRow(query, userId)

	if err := row.Scan(&count); err != nil {
		log.Errorw("error with scanning row", zap.Error(err))
		return false, err
	}

	return count == 1, nil
}
