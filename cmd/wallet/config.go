package main

import (
	"errors"
	"fmt"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"os"
	"strconv"
)

type RepositoryConf struct {
	RepositoryType repository.RepositoryType
	Host           string
	Port           int
	DBName         string
	User           string
	Password       string
}

type IdempotencyConfig struct {
	Address  string
	Password string
}

type Config struct {
	BindAddr                   string
	GracefulShutdownTimeoutSec int
	Repository                 RepositoryConf
	Idempotency                IdempotencyConfig
}

func validateConfig(c Config) error {
	if c.BindAddr == "" {
		return errors.New("unexpected `BIND_ADDR` env variable")
	}
	if c.GracefulShutdownTimeoutSec < -1 {
		return errors.New("unexpected `GRACEFUL_SHUTDOWN_TIMEOUT_SEC` env variable")
	}
	err := validateRepositoryConfig(c.Repository)
	if err != nil {
		return fmt.Errorf("unexpected repository config: %w", err)
	}
	err = validateIdempotencyConfig(c.Idempotency)
	if err != nil {
		return fmt.Errorf("unexpected idempotency config: %w", err)
	}
	return nil
}

func validateRepositoryConfig(c RepositoryConf) error {
	if !repository.ValidRepositoryType(c.RepositoryType) {
		return errors.New("unexpected `DB_TYPE` env variable")
	}
	if c.Host == "" {
		return errors.New("unexpected `DB_HOST` env variable")
	}
	if c.Port <= 0 {
		return errors.New("unexpected `DB_PORT` env variable")
	}
	if c.DBName == "" {
		return errors.New("unexpected `DB_DATABASE_NAME` env variable")
	}
	if c.User == "" {
		return errors.New("unexpected `DB_USER` env variable")
	}
	if c.Password == "" {
		return errors.New("unexpected `DB_PASSWORD` env variable")
	}
	return nil
}

func validateIdempotencyConfig(c IdempotencyConfig) error {
	if c.Address == "" {
		return errors.New("unexpected `IDEMPOTENCY_REDIS_ADDR` env variable")
	}
	return nil
}

func loadConfigFromEnv() Config {
	return Config{
		BindAddr: os.Getenv("BIND_ADDR"),
		GracefulShutdownTimeoutSec: func() int {
			const defaultTimeout = 15
			raw := os.Getenv("GRACEFUL_SHUTDOWN_TIMEOUT_SEC")
			if raw == "" {
				return defaultTimeout
			}
			p, err := strconv.Atoi(raw)
			if err != nil {
				panic("cannot parse ``; " + err.Error())
			}
			return p
		}(),
		Repository: RepositoryConf{
			RepositoryType: repository.RepositoryType(os.Getenv("DB_TYPE")),
			Host:           os.Getenv("DB_HOST"),
			Port: func() int {
				p, err := strconv.Atoi(os.Getenv("DB_PORT"))
				if err != nil {
					panic("cannot parse `DB_PORT`; " + err.Error())
				}
				return p
			}(),
			DBName:   os.Getenv("DB_DATABASE_NAME"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
		},
		Idempotency: IdempotencyConfig{
			Address:  os.Getenv("IDEMPOTENCY_REDIS_ADDR"),
			Password: os.Getenv("IDEMPOTENCY_REDIS_PASSWORD"),
		},
	}
}

func newRepositoryConfigFrom(config RepositoryConf) repository.RepositoryConfig {
	r := new(repository.RepositoryConfig).
		SetRepositoryType(config.RepositoryType).
		SetHost(config.Host).
		SetPort(config.Port).
		SetDbName(config.DBName).
		SetUser(config.User).
		SetPassword(config.Password)
	return *r
}
