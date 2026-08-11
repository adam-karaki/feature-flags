package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"feature-flags/internal/flags"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func New(ctx context.Context, url string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() { r.pool.Close() }

func (r *Repository) Create(ctx context.Context, flag *flags.Flag) error {
	now := time.Now().UTC()
	flag.Version = 1
	flag.CreatedAt, flag.UpdatedAt = now, now
	rules, err := json.Marshal(flag.Rules)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO flags(name, enabled, rules, version, created_at, updated_at) VALUES($1,$2,$3,$4,$5,$6)`, flag.Name, flag.Enabled, rules, flag.Version, flag.CreatedAt, flag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create flag: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, name string) (*flags.Flag, error) {
	var f flags.Flag
	var rules []byte
	err := r.pool.QueryRow(ctx, `SELECT name, enabled, rules, version, created_at, updated_at FROM flags WHERE name=$1`, name).Scan(&f.Name, &f.Enabled, &rules, &f.Version, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, flags.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get flag: %w", err)
	}
	if err := json.Unmarshal(rules, &f.Rules); err != nil {
		return nil, fmt.Errorf("decode rules: %w", err)
	}
	return &f, nil
}

func (r *Repository) Update(ctx context.Context, flag *flags.Flag, expectedVersion int64) error {
	rules, err := json.Marshal(flag.Rules)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	newVersion := expectedVersion + 1
	result, err := r.pool.Exec(ctx, `UPDATE flags SET enabled=$1, rules=$2, version=$3, updated_at=$4 WHERE name=$5 AND version=$6`, flag.Enabled, rules, newVersion, now, flag.Name, expectedVersion)
	if err != nil {
		return fmt.Errorf("update flag: %w", err)
	}
	if result.RowsAffected() == 0 {
		if _, err := r.Get(ctx, flag.Name); errors.Is(err, flags.ErrNotFound) {
			return flags.ErrNotFound
		}
		return flags.ErrConflict
	}
	flag.Version, flag.UpdatedAt = newVersion, now
	return nil
}

func (r *Repository) Delete(ctx context.Context, name string, expectedVersion int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM flags WHERE name=$1 AND version=$2`, name, expectedVersion)
	if err != nil {
		return fmt.Errorf("delete flag: %w", err)
	}
	if result.RowsAffected() == 0 {
		if _, err := r.Get(ctx, name); errors.Is(err, flags.ErrNotFound) {
			return flags.ErrNotFound
		}
		return flags.ErrConflict
	}
	return nil
}
