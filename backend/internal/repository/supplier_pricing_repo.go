package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/supplierpricing"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type supplierPricingRepository struct {
	client *ent.Client
}

func NewSupplierPricingRepository(client *ent.Client) service.SupplierPricingRepository {
	return &supplierPricingRepository{client: client}
}

func (r *supplierPricingRepository) ListActiveByModel(ctx context.Context, modelID string, at time.Time) ([]service.SupplierPricing, error) {
	rows, err := r.client.SupplierPricing.Query().
		Where(
			supplierpricing.ModelID(modelID),
			supplierpricing.AvailabilityEQ(service.StatusActive),
			supplierpricing.EffectiveFromLTE(at),
			supplierpricing.Or(
				supplierpricing.EffectiveToIsNil(),
				supplierpricing.EffectiveToGT(at),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []service.SupplierPricing{}, nil
	}
	out := make([]service.SupplierPricing, 0, len(rows))
	for _, r := range rows {
		sp := service.SupplierPricing{
			ModelID:      derefString(r.ModelID),
			SupplierCode: r.SupplierCode,
			Version:      r.Version,
			Availability: service.StatusActive,
		}
		if r.TierLabel != nil {
			sp.TierLabel = *r.TierLabel
		}
		if r.InputRate != nil {
			sp.InputRate = decimal.NewFromFloat(*r.InputRate)
		}
		if r.OutputRate != nil {
			sp.OutputRate = decimal.NewFromFloat(*r.OutputRate)
		}
		if r.CachedReadRate != nil {
			sp.CachedReadRate = decimal.NewFromFloat(*r.CachedReadRate)
		}
		if r.CachedWriteRate != nil {
			sp.CachedWriteRate = decimal.NewFromFloat(*r.CachedWriteRate)
		}
		sp.EffectiveFrom = r.EffectiveFrom
		sp.EffectiveTo = r.EffectiveTo
		out = append(out, sp)
	}
	return out, nil
}
