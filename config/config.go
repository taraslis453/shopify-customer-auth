package config

import "time"

type (
	Config struct {
		HTTP
		Log
		Auth
		PostgreSQL
	}

	HTTP struct {
		Port string `env:"HTTP_PORT" env-default:"8080"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL" env-default:"debug"`
	}

	Auth struct {
		TokenIssuer          string        `env:"AUTH_TOKEN_ISSUER"                    env-default:"API"`
		TokenSecretKey       string        `env:"AUTH_TOKEN_SECRET_KEY"                env-default:"2fg6wuCkkQ4HNjCo"`
		AccessTokenLifetime  time.Duration `env:"AUTH_ACCESS_TOKEN_LIFETIME"           env-default:"1h"`
		RefreshTokenLifetime time.Duration `env:"AUTH_REFRESH_TOKEN_LIFETIME"          env-default:"24h"`
	}

	PostgreSQL struct {
		User     string `env:"POSTGRESQL_USER" env-default:"postgres"`
		Password string `env:"POSTGRESQL_PASSWORD" env-default:"postgres"`
		Host     string `env:"POSTGRESQL_HOST" env-default:"localhost"`
		Database string `env:"POSTGRESQL_DATABASE" env-default:"api"`
	}
)
