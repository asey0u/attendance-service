package appsettings

import (
	"context"
	"database/sql"
	"errors"

	"github.com/asey0u/attendance-service/internal/database"
)

type Repository struct {
	db database.DBTX
}

func NewRepository(db database.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbtx(ctx context.Context) database.DBTX {
	if db := database.FromContext(ctx); db != nil {
		return db
	}
	return r.db
}

func (r *Repository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.dbtx(ctx).QueryRowContext(ctx,
		`SELECT value FROM settings WHERE key = $1`, key,
	).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

func (r *Repository) Set(ctx context.Context, key, value string) error {
	_, err := r.dbtx(ctx).ExecContext(ctx,
		`INSERT INTO settings(key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		key, value,
	)
	return err
}
