package repository

import (
	"context"
	"database/sql"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/modelbalance"
	"github.com/Wei-Shaw/sub2api/ent/supplierpricing"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

type adminOperationalRepository struct {
	db     *sql.DB
	client *dbent.Client
}

func NewAdminOperationalRepository(db *sql.DB, client *dbent.Client) service.AdminOperationalRepository {
	return &adminOperationalRepository{db: db, client: client}
}

func (r *adminOperationalRepository) GetModelMargins(ctx context.Context, start, end time.Time) ([]service.AdminModelMargin, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT model, COALESCE(SUM(display_total_cost), 0), COALESCE(SUM(cost_total), 0),
			COALESCE(cost_supplier_code, ''), pricing_version
		FROM usage_logs WHERE created_at >= $1 AND created_at < $2
		GROUP BY model, cost_supplier_code, pricing_version ORDER BY model`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byModel := make(map[string]*service.AdminModelMargin)
	order := make([]string, 0)
	for rows.Next() {
		var model, supplier string
		var display, cost float64
		var version int
		if err := rows.Scan(&model, &display, &cost, &supplier, &version); err != nil {
			return nil, err
		}
		item, ok := byModel[model]
		if !ok {
			item = &service.AdminModelMargin{Model: model, SupplierBreakdown: []service.AdminSupplierBreakdown{}, PricingVersions: []int{}}
			byModel[model] = item
			order = append(order, model)
		}
		item.DisplayTotal += display
		item.CostTotal += cost
		item.SupplierBreakdown = append(item.SupplierBreakdown, service.AdminSupplierBreakdown{SupplierCode: supplier, DisplayTotal: display, CostTotal: cost, Margin: display - cost})
		item.PricingVersions = appendUniqueInt(item.PricingVersions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]service.AdminModelMargin, 0, len(order))
	for _, model := range order {
		item := *byModel[model]
		item.Margin = item.DisplayTotal - item.CostTotal
		if item.DisplayTotal != 0 {
			item.MarginPercent = item.Margin / item.DisplayTotal * 100
		}
		out = append(out, item)
	}
	return out, nil
}

func appendUniqueInt(values []int, value int) []int {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func (r *adminOperationalRepository) GetUserModelBalances(ctx context.Context, userID int64) ([]service.AdminModelBalance, error) {
	rows, err := r.client.ModelBalance.Query().Where(modelbalance.UserIDEQ(userID)).WithModel().All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.AdminModelBalance, 0, len(rows))
	for _, row := range rows {
		canonical := row.ModelID
		if row.Edges.Model != nil {
			canonical = row.Edges.Model.CanonicalName
		}
		status := row.Status
		if status == "" {
			status = "not_purchased"
		}
		out = append(out, service.AdminModelBalance{UserID: row.UserID, ModelID: row.ModelID, CanonicalName: canonical, TokensPurchased: row.TokensPurchased, TokensConsumed: row.TokensConsumed, Balance: row.Balance, UsagePercent: row.UsagePercent, Status: status})
	}
	return out, nil
}

func (r *adminOperationalRepository) GetSupplierPricing(ctx context.Context, model string, at time.Time) ([]service.SupplierPricing, error) {
	rows, err := r.client.SupplierPricing.Query().Where(supplierpricing.ModelIDEQ(model), supplierpricing.AvailabilityEQ(service.StatusActive), supplierpricing.EffectiveFromLTE(at), supplierpricing.Or(supplierpricing.EffectiveToIsNil(), supplierpricing.EffectiveToGT(at))).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.SupplierPricing, 0, len(rows))
	for _, row := range rows {
		item := service.SupplierPricing{SupplierCode: row.SupplierCode, Version: row.Version, EffectiveFrom: row.EffectiveFrom, EffectiveTo: row.EffectiveTo}
		if row.ModelID != nil {
			item.ModelID = *row.ModelID
		}
		if row.Availability != nil {
			item.Availability = *row.Availability
		}
		if row.InputRate != nil {
			item.InputRate = decimal.NewFromFloat(*row.InputRate)
		}
		if row.OutputRate != nil {
			item.OutputRate = decimal.NewFromFloat(*row.OutputRate)
		}
		if row.CachedReadRate != nil {
			item.CachedReadRate = decimal.NewFromFloat(*row.CachedReadRate)
		}
		if row.CachedWriteRate != nil {
			item.CachedWriteRate = decimal.NewFromFloat(*row.CachedWriteRate)
		}
		if row.TierLabel != nil {
			item.TierLabel = *row.TierLabel
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *adminOperationalRepository) GetReconciliationDrift(ctx context.Context, start, end time.Time) (*service.AdminReconciliationDrift, error) {
	result := &service.AdminReconciliationDrift{}
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(display_total_cost), 0) FROM usage_logs WHERE created_at >= $1 AND created_at < $2`, start, end).Scan(&result.UsageLogsCount, &result.UsageLogsDisplay)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(display_total_cost), 0) FROM usage_model_snapshots WHERE updated_at >= $1 AND updated_at < $2`, start, end).Scan(&result.SnapshotsCount, &result.SnapshotsDisplay)
	if err != nil {
		return nil, err
	}
	result.DisplayTotalDrift = result.UsageLogsDisplay - result.SnapshotsDisplay
	return result, nil
}

var _ service.AdminOperationalRepository = (*adminOperationalRepository)(nil)
