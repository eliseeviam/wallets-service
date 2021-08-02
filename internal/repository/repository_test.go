package repository_test

import (
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidRepositoryType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.True(t, repository.ValidRepositoryType(repository.RepositoryPosgreSQL))
		require.True(t, repository.ValidRepositoryType(repository.RepositoryMySQL))
	})

	t.Run("invelid", func(t *testing.T) {
		require.False(t, repository.ValidRepositoryType("unknown_repo"))
	})
}

func TestNewWalletsRepository(t *testing.T) {

	t.Run("not found", func(t *testing.T) {
		r := new(repository.RepositoryConfig).
			SetRepositoryType("unknown_repo")

		repo, err := repository.NewWalletsCreatorRepository(*r)
		require.Equal(t, repository.ErrUnknownRepositoryType, err)
		require.Nil(t, repo)
	})
}
