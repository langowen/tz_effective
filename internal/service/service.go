package service

import (
	"context"
	"log/slog"
	"tz_effective/deploy/config"
	"tz_effective/internal/entities"
)

type Service struct {
	storage Storage
	cfg     *config.Config
}

func NewService(storage Storage, cfg *config.Config) *Service {
	return &Service{
		storage: storage,
		cfg:     cfg,
	}
}

func (s *Service) CreateSubscription(ctx context.Context, sub *entities.Subscriptions) (int64, error) {
	return s.storage.CreateSubscription(ctx, sub)
}

func (s *Service) GetSubscription(ctx context.Context, id int64) (*entities.Subscriptions, error) {
	return s.storage.GetSubscription(ctx, id)
}

func (s *Service) UpdateSubscription(ctx context.Context, id int64, sub *entities.Subscriptions) error {
	return s.storage.UpdateSubscription(ctx, id, sub)
}

func (s *Service) DeleteSubscription(ctx context.Context, id int64) error {
	return s.storage.DeleteSubscription(ctx, id)
}

func (s *Service) ListSubscriptions(ctx context.Context, filter *entities.ListFilter) ([]entities.Subscriptions, error) {
	return s.storage.ListSubscriptions(ctx, filter)
}

func (s *Service) CalculateTotalCost(ctx context.Context, filter *entities.CostFilter) (int64, error) {
	slog.Info("Calculating total cost", "filter", filter)
	return s.storage.CalculateTotalCost(ctx, filter)
}
