package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"testAlvtoShp/internal/logger"
)

func TestInitConfig_Success(t *testing.T) {
	err := os.Setenv("DATABASE_PORT", "5432")
	assert.NoError(t, err)
	err = os.Setenv("DATABASE_HOST", "localhost")
	assert.NoError(t, err)
	err = os.Setenv("DATABASE_USER", "testuser")
	assert.NoError(t, err)
	err = os.Setenv("DATABASE_PASSWORD", "testpass")
	assert.NoError(t, err)
	err = os.Setenv("DATABASE_NAME", "testdb")
	assert.NoError(t, err)
	err = os.Setenv("JWT_SECRET", "mysecret")
	assert.NoError(t, err)
	err = os.Setenv("SERVER_PORT", "8080")
	assert.NoError(t, err)
	defer func() {
		err = os.Unsetenv("DATABASE_PORT")
		assert.NoError(t, err)
	}()
	defer func() {
		err = os.Unsetenv("DATABASE_HOST")
		assert.NoError(t, err)
	}()
	defer func() {
		err = os.Unsetenv("DATABASE_USER")
		assert.NoError(t, err)
	}()
	defer func() {
		err = os.Unsetenv("DATABASE_PASSWORD")
		assert.NoError(t, err)
	}()
	defer func() {
		err = os.Unsetenv("DATABASE_NAME")
		assert.NoError(t, err)
	}()
	defer func() {
		err = os.Unsetenv("JWT_SECRET")
		assert.NoError(t, err)
	}()
	defer func() {
		err = os.Unsetenv("SERVER_PORT")
		assert.NoError(t, err)
	}()

	ctx := context.Background()
	ctx = logger.ContextWithLogger(ctx, logger.InitLogger())

	cfg := InitConfig(ctx)

	assert.NotNil(t, cfg, "Config should not be nil")
	assert.Equal(t, "5432", cfg.DbPort)
	assert.Equal(t, "localhost", cfg.DbHost)
	assert.Equal(t, "testuser", cfg.DbUser)
	assert.Equal(t, "testpass", cfg.DbPass)
	assert.Equal(t, "testdb", cfg.DbName)
	assert.Equal(t, "mysecret", cfg.JwtSecret)
	assert.Equal(t, "8080", cfg.Port)
}
