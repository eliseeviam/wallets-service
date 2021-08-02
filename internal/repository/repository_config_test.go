package repository_test

import (
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRepositoryConfig(t *testing.T) {

	const (
		expectedRepositoryType = repository.RepositoryPosgreSQL
		expectedHost           = "test_host"
		expectedPort           = 1111
		expectedDbName         = "test_db"
		expectedUser           = "test_user"
		expectedPassword       = "test_password"
	)

	r := new(repository.RepositoryConfig).
		SetRepositoryType(expectedRepositoryType).
		SetHost(expectedHost).
		SetPort(expectedPort).
		SetDbName(expectedDbName).
		SetUser(expectedUser).
		SetPassword(expectedPassword)

	require.Equal(t, expectedRepositoryType, r.RepositoryType())
	require.Equal(t, expectedHost, r.Host())
	require.Equal(t, expectedPort, r.Port())
	require.Equal(t, expectedDbName, r.DBName())
	require.Equal(t, expectedUser, r.User())
	require.Equal(t, expectedPassword, r.Password())

}
