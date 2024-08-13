package config

import "time"

const (
	engine = "postgres"

	defaultMaxIdleConns    = 5
	defaultMaxOpenConns    = 7
	defaultConnMaxLifetime = 30 * time.Minute
)

func LoadPostgres() Database {
	cfg := Database{
		Leader: DatabaseConfig{
			Host:            RequireEnv("POSTGRES_LEADER_HOSTNAME"),
			Port:            RequireEnv("POSTGRES_LEADER_PORT"),
			Username:        RequireEnv("POSTGRES_LEADER_USERNAME"),
			Password:        RequireEnv("POSTGRES_LEADER_PASSWORD"),
			DB:              RequireEnv("POSTGRES_LEADER_DATABASE_NAME"),
			Scheme:          engine,
			MaxIdleConns:    RequireEnvToInt("POSTGRES_LEADER_MAX_IDLE_CONNECTIONS"),
			MaxOpenConns:    RequireEnvToInt("POSTGRES_LEADER_MAX_OPEN_CONNECTIONS"),
			ConnMaxLifetime: RequireEnvToDuration("POSTGRES_LEADER_CONNECTION_MAX_LIFETIME"),
		},
		Follower: DatabaseConfig{
			Host:            OptionalEnv("POSTGRES_FOLLOWER_HOSTNAME", ""),
			Port:            OptionalEnv("POSTGRES_FOLLOWER_PORT", ""),
			Username:        OptionalEnv("POSTGRES_FOLLOWER_USERNAME", ""),
			Password:        OptionalEnv("POSTGRES_FOLLOWER_PASSWORD", ""),
			DB:              OptionalEnv("POSTGRES_FOLLOWER_DATABASE_NAME", ""),
			Scheme:          engine,
			MaxIdleConns:    OptionalEnvToInt("POSTGRES_FOLLOWER_MAX_IDLE_CONNECTIONS", defaultMaxIdleConns),
			MaxOpenConns:    OptionalEnvToInt("POSTGRES_FOLLOWER_MAX_OPEN_CONNECTIONS", defaultMaxOpenConns),
			ConnMaxLifetime: OptionalEnvToDuration("POSTGRES_FOLLOWER_CONNECTION_MAX_LIFETIME", defaultConnMaxLifetime),
		},
	}

	return cfg
}
