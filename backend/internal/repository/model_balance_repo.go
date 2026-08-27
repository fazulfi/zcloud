package repository

import (
	"context"
	"errors"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/modelbalance"
	"github.com/Wei-Shaw/sub2api/ent/modelcatalog"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type modelBalanceRepository struct {
	client *dbent.Client
}

// NewModelBalanceRepository creates an Ent-backed model balance repository.
func NewModelBalanceRepository(client *dbent.Client) service.ModelBalanceRepository {
	return &modelBalanceRepository{client: client}
}

func (r *modelBalanceRepository) GetByUserAndModel(ctx context.Context, userID int64, modelName string) (*service.ModelBalance, error) {
	if r == nil || r.client == nil {
		return nil, errors.New("model balance repository client is nil")
	}
	row, err := r.client.ModelBalance.Query().
		Where(
			modelbalance.UserIDEQ(userID),
			modelbalance.HasModelWith(modelcatalog.CanonicalName(modelName)),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &service.ModelBalance{UserID: row.UserID, ModelID: row.ModelID, Balance: row.Balance, Status: row.Status}, nil
}

func (r *modelBalanceRepository) SetBlocked(ctx context.Context, userID int64, modelID string, blocked bool) error {
	if r == nil || r.client == nil {
		return errors.New("model balance repository client is nil")
	}
	status := "active"
	if blocked {
		status = "blocked"
	}
	_, err := r.client.ModelBalance.Update().
		Where(modelbalance.UserIDEQ(userID), modelbalance.ModelIDEQ(modelID)).
		SetStatus(status).
		Save(ctx)
	return err
}

var _ service.ModelBalanceRepository = (*modelBalanceRepository)(nil)
