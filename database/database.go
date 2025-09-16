package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Domedik/trussrod/settings"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func getURL(c *settings.DatabaseConfig) string {
	var userInfo *url.Userinfo
	if c.Password != "" {
		userInfo = url.UserPassword(c.User, c.Password)
	} else {
		userInfo = url.User(c.User)
	}
	var driver = "postgres"
	if c.Driver != "" {
		driver = c.Driver
	}

	u := &url.URL{
		Scheme: driver,
		User:   userInfo,
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   c.Name,
	}

	q := url.Values{}
	if c.SSLMode != "" {
		q.Set("sslmode", c.SSLMode)
	}
	if c.SearchPath != "" {
		q.Set("options", fmt.Sprintf("-c search_path=%s", c.SearchPath))
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func New(c *settings.DatabaseConfig) (*DB, error) {
	var err error
	cfg, err := pgxpool.ParseConfig(getURL(c))
	if err != nil {
		return nil, err
	}
	cfg.MinConns = 2
	cfg.MaxConns = 20
	cfg.MaxConnLifetime = 10 * time.Minute
	cfg.MaxConnIdleTime = 20 * time.Minute
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)

	return &DB{
		Pool: pool,
	}, err
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
