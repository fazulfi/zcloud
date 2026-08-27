package service

import (
	"context"
	"time"
)

// AdminOperationalRepository provides admin-only operational evidence.
type AdminOperationalRepository interface {
	GetModelMargins(ctx context.Context, start, end time.Time) ([]AdminModelMargin, error)
	GetUserModelBalances(ctx context.Context, userID int64) ([]AdminModelBalance, error)
	GetSupplierPricing(ctx context.Context, model string, at time.Time) ([]SupplierPricing, error)
	GetReconciliationDrift(ctx context.Context, start, end time.Time) (*AdminReconciliationDrift, error)
}

type AdminSupplierBreakdown struct {
	SupplierCode string  `json:"supplier_code"`
	DisplayTotal float64 `json:"display_total"`
	CostTotal    float64 `json:"cost_total"`
	Margin       float64 `json:"margin"`
}

type AdminModelMargin struct {
	Model             string                   `json:"model"`
	DisplayTotal      float64                  `json:"display_total"`
	CostTotal         float64                  `json:"cost_total"`
	Margin            float64                  `json:"margin"`
	MarginPercent     float64                  `json:"margin_percent"`
	SupplierBreakdown []AdminSupplierBreakdown `json:"supplier_breakdown"`
	PricingVersions   []int                    `json:"pricing_versions"`
}

type AdminModelBalance struct {
	UserID          int64   `json:"user_id"`
	ModelID         string  `json:"model_id"`
	CanonicalName   string  `json:"canonical_name"`
	TokensPurchased int64   `json:"tokens_purchased"`
	TokensConsumed  int64   `json:"tokens_consumed"`
	Balance         int64   `json:"balance"`
	UsagePercent    float64 `json:"usage_percent"`
	Status          string  `json:"status"`
}

type AdminReconciliationDrift struct {
	UsageLogsCount    int64   `json:"usage_logs_count"`
	SnapshotsCount    int64   `json:"snapshots_count"`
	UsageLogsDisplay  float64 `json:"usage_logs_display_total"`
	SnapshotsDisplay  float64 `json:"snapshots_display_total"`
	DisplayTotalDrift float64 `json:"display_total_drift"`
}
