package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/Domedik/trussrod/settings"
	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sql.DB
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
	dsn := getURL(c)
	var err error
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(50)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(10 * time.Minute)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	return &DB{
		Conn: conn,
	}, nil
}

func (db *DB) Close() {
	db.Conn.Close()
}

func (db *DB) Prepare(ctx context.Context, query string) (*sql.Stmt, error) {
	return db.Conn.PrepareContext(ctx, query)
}

func (db *DB) Ping(ctx context.Context) error {
	return db.Conn.PingContext(ctx)
}
