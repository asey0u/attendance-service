package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type ctxKey struct{}

func BeginTx(ctx context.Context, db DBTX) (*sql.Tx, error) {
	type beginner interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	}
	if b, ok := db.(beginner); ok {
		return b.BeginTx(ctx, nil)
	}
	return nil, fmt.Errorf("db does not support transactions")
}

func WithDB(ctx context.Context, db DBTX) context.Context {
	return context.WithValue(ctx, ctxKey{}, db)
}

func FromContext(ctx context.Context) DBTX {
	db, _ := ctx.Value(ctxKey{}).(DBTX)
	return db
}

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	deadline := time.Now().Add(20 * time.Second)
	for {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("ping postgres: %w", err)
		}
		time.Sleep(time.Second)
	}
}
