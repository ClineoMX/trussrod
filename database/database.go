package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/Domedik/trussrod/settings"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
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
	q.Set("sslmode", c.SSLMode) // 'disable' is the default value

	if c.SearchPath != "" {
		q.Set("options", fmt.Sprintf("-c search_path=%s", c.SearchPath))
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func NewPostgres(c *settings.DatabaseConfig) (*Postgres, error) {
	var err error
	cfg, err := pgxpool.ParseConfig(getURL(c))
	if err != nil {
		return nil, err
	}
	cfg.MinConns = 0
	if c.MaxConns > 0 {
		cfg.MaxConns = int32(c.MaxConns)
	} else {
		if os.Getenv("IS_ASYNC") != "" {
			cfg.MaxConns = 2
		} else {
			cfg.MaxConns = 20
		}
	}
	cfg.MaxConnLifetime = 10 * time.Minute
	cfg.MaxConnIdleTime = 20 * time.Minute
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		Pool: pool,
	}, nil
}

func (db *Postgres) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return db.Pool.Query(ctx, sql, args...)
}

func (db *Postgres) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return db.Pool.QueryRow(ctx, sql, args...)
}

func (db *Postgres) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	return db.Pool.Exec(ctx, sql, args...)
}

func (db *Postgres) Close() {
	db.Pool.Close()
}

func (db *Postgres) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *Postgres) BeginTx(ctx context.Context, opts any) (Tx, error) {
	options, ok := opts.(pgx.TxOptions)
	if !ok {
		return nil, fmt.Errorf("invalid transaction options")
	}

	return db.Pool.BeginTx(ctx, options)
}
