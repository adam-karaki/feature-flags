package flags

import (
	"context"
	"fmt"
)

type Service struct {
	repo  Repository
	cache Cache
	pub   EventPublisher
}

func NewService(repo Repository, cache Cache, pub EventPublisher) *Service {
	return &Service{repo: repo, cache: cache, pub: pub}
}

func (s *Service) Create(ctx context.Context, flag *Flag) error {
	if err := ValidateName(flag.Name); err != nil {
		return err
	}
	for _, rule := range flag.Rules {
		if err := ValidateRule(rule); err != nil {
			return err
		}
	}
	if err := s.repo.Create(ctx, flag); err != nil {
		return err
	}
	if err := s.cache.Set(ctx, flag); err != nil {
		return fmt.Errorf("cache flag: %w", err)
	}
	return s.publish(ctx, Event{Type: "created", Flag: *flag, Version: flag.Version})
}

func (s *Service) Get(ctx context.Context, name string) (*Flag, error) {
	if flag, err := s.cache.Get(ctx, name); err == nil && flag != nil {
		return flag, nil
	}
	flag, err := s.repo.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Set(ctx, flag)
	return flag, nil
}

func (s *Service) Update(ctx context.Context, flag *Flag, expectedVersion int64) error {
	if err := ValidateName(flag.Name); err != nil {
		return err
	}
	for _, rule := range flag.Rules {
		if err := ValidateRule(rule); err != nil {
			return err
		}
	}
	if err := s.repo.Update(ctx, flag, expectedVersion); err != nil {
		return err
	}
	if err := s.cache.Set(ctx, flag); err != nil {
		return fmt.Errorf("cache flag: %w", err)
	}
	return s.publish(ctx, Event{Type: "updated", Flag: *flag, Version: flag.Version})
}

func (s *Service) Delete(ctx context.Context, name string, expectedVersion int64) error {
	if err := s.repo.Delete(ctx, name, expectedVersion); err != nil {
		return err
	}
	if err := s.cache.Delete(ctx, name); err != nil {
		return fmt.Errorf("delete cache: %w", err)
	}
	return s.publish(ctx, Event{Type: "deleted", Flag: Flag{Name: name}, Version: expectedVersion + 1})
}

func (s *Service) Evaluate(ctx context.Context, name string, input EvaluationContext) (EvaluationResult, error) {
	flag, err := s.Get(ctx, name)
	if err != nil {
		return EvaluationResult{}, err
	}
	return Evaluate(flag, input), nil
}

func (s *Service) publish(ctx context.Context, event Event) error {
	if err := s.pub.Publish(ctx, event); err != nil {
		// The database/cache mutation is already durable. Event publication is
		// deliberately returned to the caller so failures are visible rather
		// than silently pretending propagation succeeded.
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}
