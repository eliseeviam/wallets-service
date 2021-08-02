package main

import (
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		c := Config{
			BindAddr:                   ":8080",
			GracefulShutdownTimeoutSec: -1,
			Repository: RepositoryConf{
				RepositoryType: repository.RepositoryPosgreSQL,
				Host:           "localhost",
				Port:           5432,
				DBName:         "test",
				User:           "user",
				Password:       "pwd",
			},
			Idempotency: IdempotencyConfig{
				Address:  "localhost:6379",
				Password: "",
			},
		}
		require.NoError(t, validateConfig(c))
	})

	t.Run("notOk", func(t *testing.T) {
		c := Config{
			BindAddr:                   "",
			GracefulShutdownTimeoutSec: -1,
			Repository: RepositoryConf{
				RepositoryType: repository.RepositoryPosgreSQL,
				Host:           "localhost",
				Port:           5432,
				DBName:         "test",
				User:           "user",
				Password:       "pwd",
			},
		}
		require.Error(t, validateConfig(c))
	})

	t.Run("notOk", func(t *testing.T) {
		c := Config{
			BindAddr:                   ":8080",
			GracefulShutdownTimeoutSec: -2,
			Repository: RepositoryConf{
				RepositoryType: repository.RepositoryPosgreSQL,
				Host:           "localhost",
				Port:           5432,
				DBName:         "test",
				User:           "user",
				Password:       "pwd",
			},
		}
		require.Error(t, validateConfig(c))
	})
}
