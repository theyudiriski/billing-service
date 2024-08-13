package config

import (
	"time"
)

func LoadAPI() API {
	config := API{}

	config.HTTP.Port = RequireEnvToInt("HTTP_PORT")
	config.HTTP.ReadTimeout = RequireEnvToDuration("HTTP_READ_TIMEOUT")
	config.HTTP.WriteTimeout = RequireEnvToDuration("HTTP_WRITE_TIMEOUT")

	config.Database = LoadPostgres()

	return config
}

type API struct {
	HTTP struct {
		Port         int
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}
	Database Database
}
