package repository

import (
	"context"

	"entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type usageModelSnapshotRepository struct {
	client *dbent.Client
}

func NewUsageModelSnapshotRepository(client *dbent.Client) service.UsageModelSnapshotRepository {
	return &usageModelSnapshotRepository{client: client}
}

func (r *usageModelSnapshotRepository) Upsert(ctx context.Context, snap *service.UsageModelSnapshot) error {
	if r == nil || r.client == nil {
		return nil
	}
	if snap == nil || snap.UserID <= 0 || snap.Model == "" {
		return nil
	}
	create := r.client.UsageModelSnapshot.Create().
		SetUserID(snap.UserID).
		SetModel(snap.Model).
		SetPricingVersion(snap.PricingVersion).
		SetDisplayInputCost(snap.DisplayInputCost).
		SetDisplayOutputCost(snap.DisplayOutputCost).
		SetDisplayCacheReadCost(snap.DisplayCacheReadCost).
		SetDisplayCacheWriteCost(snap.DisplayCacheWriteCost).
		SetDisplayTotalCost(snap.DisplayTotalCost).
		SetDisplayBlendCost(snap.DisplayBlendCost).
		SetCostInput(snap.CostInput).
		SetCostOutput(snap.CostOutput).
		SetCostCacheRead(snap.CostCacheRead).
		SetCostCacheWrite(snap.CostCacheWrite).
		SetCostTotal(snap.CostTotal).
		SetCostSupplierCode(snap.CostSupplierCode).
		SetInputTokens(snap.InputTokens).
		SetOutputTokens(snap.OutputTokens).
		SetCacheReadTokens(snap.CacheReadTokens).
		SetCacheWriteTokens(snap.CacheWriteTokens).
		SetUsageModelPct(snap.UsageModelPct).
		OnConflict(
			sql.ConflictColumns("user_id", "model", "pricing_version"),
		).
		UpdateNewValues().
		AddDisplayInputCost(snap.DisplayInputCost).
		AddDisplayOutputCost(snap.DisplayOutputCost).
		AddDisplayCacheReadCost(snap.DisplayCacheReadCost).
		AddDisplayCacheWriteCost(snap.DisplayCacheWriteCost).
		AddDisplayTotalCost(snap.DisplayTotalCost).
		AddDisplayBlendCost(snap.DisplayBlendCost).
		AddCostInput(snap.CostInput).
		AddCostOutput(snap.CostOutput).
		AddCostCacheRead(snap.CostCacheRead).
		AddCostCacheWrite(snap.CostCacheWrite).
		AddCostTotal(snap.CostTotal).
		SetCostSupplierCode(snap.CostSupplierCode).
		AddInputTokens(snap.InputTokens).
		AddOutputTokens(snap.OutputTokens).
		AddCacheReadTokens(snap.CacheReadTokens).
		AddCacheWriteTokens(snap.CacheWriteTokens).
		SetUsageModelPct(snap.UsageModelPct)
	return create.Exec(ctx)
}
