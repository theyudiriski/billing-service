package config

import "time"

type Database struct {
	Leader   DatabaseConfig
	Follower DatabaseConfig
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	DB              string
	Scheme          string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func (c Database) IsFollowerEnabled() bool {
	if c.Follower.Host == "" || c.Follower.Port == "" {
		return false
	}

	return c.Follower.Host != c.Leader.Host || c.Follower.Port != c.Leader.Port
}
