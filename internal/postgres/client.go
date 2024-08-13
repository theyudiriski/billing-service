package postgres

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/theyudiriski/billing-service/config"
)

type Client struct {
	Leader   *sql.DB
	Follower *sql.DB
}

func NewClient(c config.Database) (*Client, error) {
	leaderDB, err := openDB(c.Leader)
	if err != nil {
		return nil, fmt.Errorf("failed to open leader DB: %w", err)
	}

	var followerDB *sql.DB
	if c.IsFollowerEnabled() {
		followerDB, err = openDB(c.Follower)
		if err != nil {
			return nil, fmt.Errorf("failed to open follower DB: %w", err)
		}
	} else {
		followerDB = leaderDB
	}

	return &Client{
		Leader:   leaderDB,
		Follower: followerDB,
	}, nil
}

func openDB(c config.DatabaseConfig) (*sql.DB, error) {
	var db *sql.DB
	var err error

	dbURI := &url.URL{
		Scheme: c.Scheme,
		User:   url.UserPassword(c.Username, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   c.DB,
	}

	db, err = sql.Open("pgx", dbURI.String())
	if err != nil {
		return nil, fmt.Errorf("sql.Open(): %w", err)
	}

	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)

	return db, nil
}
