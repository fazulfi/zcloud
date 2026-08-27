package service

import "context"

// ModelBalanceRepository reads and updates per-user model plan state.
type CustomerModelBalanceRepository interface {
	ListByUser(ctx context.Context, userID int64) ([]ModelBalance, error)
}

type ModelBalanceRepository interface {
	GetByUserAndModel(ctx context.Context, userID int64, modelID string) (*ModelBalance, error)
	SetBlocked(ctx context.Context, userID int64, modelID string, blocked bool) error
}

// ModelBalance is the service representation of a model balance row.
type ModelBalance struct {
	UserID          int64
	ModelID         string
	ModelName       string
	TokensPurchased int64
	TokensConsumed  int64
	Balance         int64
	UsagePercent    float64
	Status          string
}
