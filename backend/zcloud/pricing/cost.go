package pricing

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// ErrUnknownPrice is returned when a required rate is missing or
// non-positive. Unknown prices must never silently become zero (kernel §4).
var ErrUnknownPrice = errors.New("pricing: unknown price")

// ErrNoCheapestSupplier is returned when no supplier has a valid rate for
// the canonical model at the given time.
var ErrNoCheapestSupplier = errors.New("pricing: no cheapest supplier")

// CostRates holds the supplier wholesale per-M token rates for a model.
//
// Rates are quoted in $ / M tokens (kernel §4.1/§4.2) and are
// internal-only (D8: never exposed to customers).
type CostRates struct {
	SupplierCode    string
	InputPerM       decimal.Decimal
	OutputPerM      decimal.Decimal
	CachedReadPerM  decimal.Decimal
	CachedWritePerM *decimal.Decimal // nil = not billable
	Availability    string
	EffectiveFrom   time.Time
	EffectiveTo     *time.Time // nil = open-ended
	PricingVersion  int
}

// CostBreakdown is the per-token supplier cost result (kernel §5).
type CostBreakdown struct {
	SupplierCode    string
	InputCost       decimal.Decimal
	OutputCost      decimal.Decimal
	CachedReadCost  decimal.Decimal
	CachedWriteCost decimal.Decimal
	TotalCost       decimal.Decimal
	BlendCost       decimal.Decimal
	Version         int
}

// ComputeCost computes the supplier cost breakdown for a request.
//
// cost$ = in×inRate + out×outRate + cachedRead×cachedReadRate +
// cachedWrite×cachedWriteRate (kernel §5). No cache discount is applied
// here; the discount is already reflected in the supplier rate itself.
func ComputeCost(in, out, cachedRead, cachedWrite int64, rates CostRates) (CostBreakdown, error) {
	if !rates.InputPerM.IsPositive() || !rates.OutputPerM.IsPositive() {
		return CostBreakdown{}, ErrUnknownPrice
	}
	if rates.SupplierCode == "" {
		return CostBreakdown{}, ErrUnknownPrice
	}

	b := CostBreakdown{
		SupplierCode: rates.SupplierCode,
		InputCost:    NormalizePerM(rates.InputPerM).Mul(decimal.NewFromInt(in)),
		OutputCost:   NormalizePerM(rates.OutputPerM).Mul(decimal.NewFromInt(out)),
		Version:      rates.PricingVersion,
	}
	if rates.CachedReadPerM.IsPositive() {
		b.CachedReadCost = NormalizePerM(rates.CachedReadPerM).Mul(decimal.NewFromInt(cachedRead))
	}
	if rates.CachedWritePerM != nil && rates.CachedWritePerM.IsPositive() {
		b.CachedWriteCost = NormalizePerM(*rates.CachedWritePerM).Mul(decimal.NewFromInt(cachedWrite))
	}
	b.TotalCost = b.InputCost.Add(b.OutputCost).Add(b.CachedReadCost).Add(b.CachedWriteCost)
	// Blend is the routing heuristic used by the scheduler: 1:1 weighted
	// input+output, matching the M1.6 cost ranking semantics.
	b.BlendCost = b.InputCost.Add(b.OutputCost)
	return b, nil
}

// ActiveCostRate reports whether a supplier rate row is usable at time t:
// availability must be "active" and the effective window must contain t.
func ActiveCostRate(r CostRates, at time.Time) bool {
	if r.Availability != "" && r.Availability != "active" {
		return false
	}
	if !r.EffectiveFrom.IsZero() && r.EffectiveFrom.After(at) {
		return false
	}
	if r.EffectiveTo != nil && !r.EffectiveTo.After(at) {
		return false
	}
	return true
}

// CheapestActiveRate returns the active supplier rate with the lowest
// blended cost (input + output per M). Rows with missing or non-positive
// rates are skipped. An empty input yields ErrNoCheapestSupplier.
func CheapestActiveRate(rates []CostRates, at time.Time) (CostRates, error) {
	var best CostRates
	found := false
	for _, r := range rates {
		if !ActiveCostRate(r, at) {
			continue
		}
		if !r.InputPerM.IsPositive() || !r.OutputPerM.IsPositive() {
			continue
		}
		if !found {
			best = r
			found = true
			continue
		}
		if r.InputPerM.Add(r.OutputPerM).Cmp(best.InputPerM.Add(best.OutputPerM)) < 0 {
			best = r
		}
	}
	if !found {
		return CostRates{}, ErrNoCheapestSupplier
	}
	return best, nil
}

// CostBlendPerToken computes cf_in × in + out (kernel §4 cost_blend)
// normalized per token, where in/out are per-M rates.
func CostBlendPerToken(cfIn decimal.Decimal, inPerM, outPerM decimal.Decimal) decimal.Decimal {
	return cfIn.Mul(inPerM).Add(outPerM).Div(perMillion)
}
