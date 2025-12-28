package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

// pgxTxWrapper wraps pgx.Tx to implement the Tx interface.
type pgxTxWrapper struct {
	tx pgx.Tx
}

func (w *pgxTxWrapper) Commit(ctx context.Context) error {
	return w.tx.Commit(ctx)
}

func (w *pgxTxWrapper) Rollback(ctx context.Context) error {
	return w.tx.Rollback(ctx)
}

func (w *pgxTxWrapper) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return w.tx.Query(ctx, sql, args...)
}

func (w *pgxTxWrapper) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return w.tx.QueryRow(ctx, sql, args...)
}

func (w *pgxTxWrapper) Exec(ctx context.Context, sql string, args ...any) (Result, error) {
	return w.tx.Exec(ctx, sql, args...)
}

func getURL(user, password, driver, host, port, name, sslmode, searchpath string) string {
	var userInfo *url.Userinfo
	if password != "" {
		userInfo = url.UserPassword(user, password)
	} else {
		userInfo = url.User(user)
	}
	if driver == "" {
		driver = "postgres"
	}

	u := &url.URL{
		Scheme: driver,
		User:   userInfo,
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   name,
	}

	q := url.Values{}
	q.Set("sslmode", sslmode) // 'disable' is the default value

	if searchpath != "" {
		q.Set("options", fmt.Sprintf("-c search_path=%s", searchpath))
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// NewPostgres creates a new Postgres client.
// user: the username for the database
// password: the password for the database
// host: the host of the database
// port: the port of the database
// name: the name of the database
// sslmode: the ssl mode for the database
// searchpath: the search path for the database
// maxConns: the maximum number of connections to the database
// returns: a new Postgres client
// error: if the database connection fails
func NewPostgres(user, password, host, port, name, sslmode, searchpath string, maxConns int) (*Postgres, error) {
	const driver = "postgres"
	var err error
	cfg, err := pgxpool.ParseConfig(getURL(user, password, driver, host, port, name, sslmode, searchpath))
	if err != nil {
		return nil, err
	}
	cfg.MinConns = 0
	if maxConns > 0 {
		cfg.MaxConns = int32(maxConns)
	} else {
		cfg.MaxConns = 20
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

	if opts == nil {
		tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return nil, err
		}
		return &pgxTxWrapper{tx: tx}, nil
	}

	options, ok := opts.(pgx.TxOptions)
	if !ok {
		return nil, fmt.Errorf("invalid transaction options")
	}

	tx, err := db.Pool.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &pgxTxWrapper{tx: tx}, nil
}
