package flags

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("flag not found")
	ErrConflict = errors.New("version conflict")
)

type Repository interface {
	Create(ctx context.Context, flag *Flag) error
	Get(ctx context.Context, name string) (*Flag, error)
	Update(ctx context.Context, flag *Flag, expectedVersion int64) error
	Delete(ctx context.Context, name string, expectedVersion int64) error
}

type Cache interface {
	Get(ctx context.Context, name string) (*Flag, error)
	Set(ctx context.Context, flag *Flag) error
	Delete(ctx context.Context, name string) error
	Close() error
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
	Close() error
}
