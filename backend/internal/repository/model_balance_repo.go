package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/modelbalance"
	"github.com/Wei-Shaw/sub2api/ent/modelcatalog"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type modelBalanceRepository struct {
	client *dbent.Client
}

// NewModelBalanceRepository creates an Ent-backed model balance repository.
func NewModelBalanceRepository(client *dbent.Client) *modelBalanceRepository {
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
	return &service.ModelBalance{UserID: row.UserID, ModelID: row.ModelID, ModelName: modelName, TokensPurchased: row.TokensPurchased, TokensConsumed: row.TokensConsumed, Balance: row.Balance, UsagePercent: row.UsagePercent, Status: row.Status}, nil
}

func (r *modelBalanceRepository) ListByUser(ctx context.Context, userID int64) ([]service.ModelBalance, error) {
	if r == nil || r.client == nil {
		return nil, errors.New("model balance repository client is nil")
	}
	rows, err := r.client.ModelBalance.Query().
		Where(modelbalance.UserIDEQ(userID)).
		WithModel().
		All(ctx)
	if err != nil {
		return nil, err
	}
	balances := make([]service.ModelBalance, 0, len(rows))
	for _, row := range rows {
		modelName := row.ModelID
		if row.Edges.Model != nil {
			modelName = row.Edges.Model.CanonicalName
		}
		balances = append(balances, service.ModelBalance{
			UserID: row.UserID, ModelID: row.ModelID, ModelName: modelName,
			TokensPurchased: row.TokensPurchased, TokensConsumed: row.TokensConsumed,
			Balance: row.Balance, UsagePercent: row.UsagePercent, Status: row.Status,
		})
	}
	return balances, nil
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

func (r *modelBalanceRepository) CreditTokens(ctx context.Context, userID int64, modelID string, tokens int64) error {
	if r == nil || r.client == nil {
		return errors.New("model balance repository client is nil")
	}
	client := clientFromContext(ctx, r.client)
	return client.ModelBalance.Create().
		SetID(uuid.NewString()).
		SetUserID(userID).
		SetModelID(modelID).
		SetTokensPurchased(tokens).
		SetTokensConsumed(0).
		SetBalance(tokens).
		SetUsagePercent(0).
		SetStatus("active").
		SetVersion(1).
		OnConflictColumns(modelbalance.FieldUserID, modelbalance.FieldModelID).
		Update(func(u *dbent.ModelBalanceUpsert) {
			u.AddTokensPurchased(tokens).
				AddBalance(tokens).
				SetStatus("active").
				AddVersion(1)
		}).
		Exec(ctx)
}

var _ service.ModelBalanceRepository = (*modelBalanceRepository)(nil)
