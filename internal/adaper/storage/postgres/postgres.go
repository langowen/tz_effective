package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
	"tz_effective/deploy/config"
	"tz_effective/internal/entities"
)

type Storage struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewStorage(pool *pgxpool.Pool, cfg *config.Config) *Storage {
	return &Storage{
		db:  pool,
		cfg: cfg,
	}
}

func New(ctx context.Context, cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.New"

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		cfg.Storage.Host,
		cfg.Storage.Port,
		cfg.Storage.User,
		cfg.Storage.Password,
		cfg.Storage.DBName,
		cfg.Storage.SSLMode,
		cfg.Storage.Schema,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: parse config failed: %w", op, err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 10 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(ctx, cfg.Storage.Timeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		slog.Error("pgxpool connect failed", "error", err)
		return nil, fmt.Errorf("%s: pgxpool connect failed: %w", op, err)
	}

	if err := pool.Ping(ctx); err != nil {
		slog.Error("pgxpool ping failed", "error", err)
		pool.Close()
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	storageBD := NewStorage(pool, cfg)

	slog.Info("PostgresSQL storage initialized successfully")
	return storageBD, nil
}

func (s *Storage) CreateSubscription(ctx context.Context, sub *entities.Subscriptions) (int64, error) {
	row := s.db.QueryRow(ctx, `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)
	var id int64
	if err := row.Scan(&id); err != nil {
		slog.Error("Failed to create subscription", "error", err)
		return 0, fmt.Errorf("error creating subscription: %w", err)
	}
	return id, nil
}

func (s *Storage) GetSubscription(ctx context.Context, id int64) (*entities.Subscriptions, error) {
	row := s.db.QueryRow(ctx, `SELECT service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1`, id)
	var sub entities.Subscriptions
	if err := row.Scan(&sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate); err != nil {
		slog.Error("Failed to get subscription", "error", err, "id", id)
		return nil, fmt.Errorf("error getting subscription with ID %d: %w", id, err)
	}
	return &sub, nil
}

func (s *Storage) UpdateSubscription(ctx context.Context, id int64, sub *entities.Subscriptions) error {
	result, err := s.db.Exec(ctx, `UPDATE subscriptions SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5 WHERE id = $6`,
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, id)
	if err != nil {
		slog.Error("Failed to update subscription", "error", err, "id", id)
		return fmt.Errorf("error updating subscription with ID %d: %w", id, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Warn("No subscription found for update", "id", id)
		return fmt.Errorf("subscription with ID %d not found", id)
	}

	return nil
}

func (s *Storage) DeleteSubscription(ctx context.Context, id int64) error {
	result, err := s.db.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		slog.Error("Failed to delete subscription", "error", err, "id", id)
		return fmt.Errorf("error deleting subscription with ID %d: %w", id, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Warn("No subscription found for deletion", "id", id)
		return fmt.Errorf("subscription with ID %d not found", id)
	}

	return nil
}

func (s *Storage) ListSubscriptions(ctx context.Context, filter *entities.ListFilter) ([]entities.Subscriptions, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE 1=1`
	params := []interface{}{}
	paramIndex := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", paramIndex)
		params = append(params, *filter.UserID)
		paramIndex++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", paramIndex)
		params = append(params, *filter.ServiceName)
		paramIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND start_date >= $%d", paramIndex)
		params = append(params, *filter.StartDate)
		paramIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND end_date <= $%d", paramIndex)
		params = append(params, *filter.EndDate)
		paramIndex++
	}

	rows, err := s.db.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []entities.Subscriptions
	for rows.Next() {
		var sub entities.Subscriptions
		var id int64
		var endDate *string

		if err := rows.Scan(&id, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &endDate); err != nil {
			return nil, err
		}

		sub.EndDate = endDate
		subs = append(subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subs, nil
}

func (s *Storage) CalculateTotalCost(ctx context.Context, filter *entities.CostFilter) (int64, error) {
	query := `
		SELECT COALESCE(SUM(price), 0) AS total_cost
		FROM subscriptions
		WHERE 1=1
		AND (
			(end_date IS NULL OR end_date >= $1)
			AND start_date <= $2
		)
	`
	params := []interface{}{filter.StartPeriod, filter.EndPeriod}
	paramIndex := 3

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", paramIndex)
		params = append(params, *filter.UserID)
		paramIndex++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", paramIndex)
		params = append(params, *filter.ServiceName)
		paramIndex++
	}

	var totalCost int64
	err := s.db.QueryRow(ctx, query, params...).Scan(&totalCost)
	if err != nil {
		slog.Error("Failed to calculate total cost", "error", err, "filter", filter)
		return 0, fmt.Errorf("error calculating total cost: %w", err)
	}

	return totalCost, nil
}
