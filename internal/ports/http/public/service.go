package public

import (
	"context"
	"tz_effective/internal/entities"
)

type Service interface {
	CreateSubscription(ctx context.Context, sub *entities.Subscriptions) (id int64, err error)
	GetSubscription(ctx context.Context, id int64) (sub *entities.Subscriptions, err error)
	UpdateSubscription(ctx context.Context, id int64, sub *entities.Subscriptions) error
	DeleteSubscription(ctx context.Context, id int64) error
	ListSubscriptions(ctx context.Context, filter *entities.ListFilter) ([]entities.Subscriptions, error)
	CalculateTotalCost(ctx context.Context, filter *entities.CostFilter) (int64, error)
}
