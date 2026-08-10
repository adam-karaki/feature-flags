package flags

import (
	"context"
	"testing"
)

type memRepo struct{ data map[string]*Flag }

func newMemRepo() *memRepo { return &memRepo{data: map[string]*Flag{}} }
func (r *memRepo) Create(_ context.Context, f *Flag) error {
	if _, ok := r.data[f.Name]; ok {
		return ErrConflict
	}
	r.data[f.Name] = f
	return nil
}
func (r *memRepo) Get(_ context.Context, n string) (*Flag, error) {
	f, ok := r.data[n]
	if !ok {
		return nil, ErrNotFound
	}
	return f, nil
}
func (r *memRepo) Update(_ context.Context, f *Flag, v int64) error {
	old, ok := r.data[f.Name]
	if !ok {
		return ErrNotFound
	}
	if old.Version != v {
		return ErrConflict
	}
	f.Version = v + 1
	r.data[f.Name] = f
	return nil
}
func (r *memRepo) Delete(_ context.Context, n string, v int64) error {
	f, ok := r.data[n]
	if !ok {
		return ErrNotFound
	}
	if f.Version != v {
		return ErrConflict
	}
	delete(r.data, n)
	return nil
}

type memCache struct{ data map[string]*Flag }

func newMemCache() *memCache { return &memCache{data: map[string]*Flag{}} }
func (c *memCache) Get(_ context.Context, n string) (*Flag, error) {
	f, ok := c.data[n]
	if !ok {
		return nil, ErrNotFound
	}
	return f, nil
}
func (c *memCache) Set(_ context.Context, f *Flag) error     { c.data[f.Name] = f; return nil }
func (c *memCache) Delete(_ context.Context, n string) error { delete(c.data, n); return nil }
func (c *memCache) Close() error                             { return nil }

type memPub struct{}

func (memPub) Publish(context.Context, Event) error { return nil }
func (memPub) Close() error                         { return nil }

func TestServiceCreateAndGet(t *testing.T) {
	r := newMemRepo()
	c := newMemCache()
	s := NewService(r, c, memPub{})
	f := &Flag{Name: "demo", Enabled: true}
	if err := s.Create(context.Background(), f); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(context.Background(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "demo" {
		t.Fatalf("got %+v", got)
	}
}
