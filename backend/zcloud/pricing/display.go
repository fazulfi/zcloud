// Package pricing implements the dual metering formulas (kernel §4).
//
// Two independent value layers are computed for every billable request:
//   - display: official retail price shown to the customer (kernel §4);
//   - cost:    supplier wholesale cost used internally (kernel §5).
//
// display$ is the only value that may be exposed to the customer (D8).
// cost$ is internal-only and must never appear in customer UI, API
// responses, invoices, or public logs.
package pricing

import (
	"time"

	"github.com/shopspring/decimal"
)

// Normalization constant: rates are quoted in $ / M tokens and must be
// divided by 1,000,000 before per-token multiplication (kernel §4).
var perMillion = decimal.NewFromInt(1_000_000)

// Kernel blending constants (kernel §4, normative).
const (
	// InFraction is the input-share weight used for the cost blend.
	InFraction = 0.9961
	// OutFraction is the output-share weight used for the cost blend.
	OutFraction = 0.0039
	// CacheDiscount is the kernel-verified automatic prompt caching
	// discount applied to cached read (kernel §5: CACHE_DISC = 0.10).
	CacheDiscount = 0.10
	// CacheHitRatio is the assumed prompt-cache hit rate for blended
	// display cost (kernel §4: HIT = 0.85).
	CacheHitRatio = 0.85
)

// BlendWeights describes how display rates are blended into a single
// effective per-token price used for plan math and meter percentages.
//
// The weights are deterministic routing heuristics, not billing
// allocations (kernel §4 Usage column is explicitly non-normative).
type BlendWeights struct {
	InputWeight  decimal.Decimal // default 1
	OutputWeight decimal.Decimal // default 1
}

// DefaultBlendWeights returns the kernel default weights (1:1).
func DefaultBlendWeights() BlendWeights {
	return BlendWeights{
		InputWeight:  decimal.NewFromInt(1),
		OutputWeight: decimal.NewFromInt(1),
	}
}

// DisplayRates holds the official per-M token rates for a model/tier.
//
// Rate fields are quoted in $ / M tokens (kernel §4.3). Missing cache
// write is represented by nil: it is only billable as zero when the
// catalog policy explicitly defines writes as free (kernel §4).
type DisplayRates struct {
	InputPerM       decimal.Decimal
	OutputPerM      decimal.Decimal
	CachedReadPerM  decimal.Decimal
	CachedWritePerM *decimal.Decimal // nil = not billable
	ContextLength   int64            // tier predicate input
	CacheMode       string           // "auto" | "none" | "explicit"
	PricingVersion  int              // snapshot version, > 0
	EffectiveFrom   time.Time
	EffectiveTo     *time.Time // nil = open-ended
}

// DisplayBreakdown is the per-token display cost result (kernel §4).
type DisplayBreakdown struct {
	InputCost       decimal.Decimal
	OutputCost      decimal.Decimal
	CachedReadCost  decimal.Decimal
	CachedWriteCost decimal.Decimal
	TotalCost       decimal.Decimal
	// BlendCost is the weighted per-request display cost (IN×blendIn +
	// OUT×blendOut with cache discount and hit ratio applied).
	BlendCost decimal.Decimal
	// Version snapshots the pricing version used for this breakdown.
	Version int
}

// NormalizePerM converts a $ / M token rate into a per-token decimal.
func NormalizePerM(ratePerM decimal.Decimal) decimal.Decimal {
	return ratePerM.Div(perMillion)
}

// BlendDisplayRate computes the blended per-token display input rate:
//
//	blendIn = 0.15 × offIn + 0.85 × offCache
//
// where offCache = offIn × (1 − CACHE_DISC) = offIn × 0.90, i.e.:
//
//	blendIn = offIn × (0.15 + 0.85 × 0.90) = offIn × 0.915
//
// The result is quoted per token (already normalized by 1M).
func BlendDisplayRate(offInPerM, offCachePerM decimal.Decimal) decimal.Decimal {
	discounted := offCachePerM.Mul(decimal.NewFromFloat(1 - CacheDiscount))
	return offInPerM.
		Mul(decimal.NewFromFloat(0.15)).
		Add(discounted.Mul(decimal.NewFromFloat(0.85))).
		Div(perMillion)
}

// ComputeDisplay computes the display cost breakdown for a request.
//
// display$ uses the official tier selected by selectDisplayRate; cache
// hit ratio is applied to the cached-read component so the blended
// number reflects the kernel HIT = 0.85 assumption.
func ComputeDisplay(in, out, cachedRead, cachedWrite int64, rates DisplayRates) (DisplayBreakdown, error) {
	// Unknown price must never silently become zero (kernel §4).
	if !rates.InputPerM.IsPositive() || !rates.OutputPerM.IsPositive() {
		return DisplayBreakdown{}, ErrUnknownPrice
	}

	blendIn := BlendDisplayRate(rates.InputPerM, rates.CachedReadPerM)

	b := DisplayBreakdown{
		InputCost:  NormalizePerM(rates.InputPerM).Mul(decimal.NewFromInt(in)),
		OutputCost: NormalizePerM(rates.OutputPerM).Mul(decimal.NewFromInt(out)),
		BlendCost: blendIn.Mul(decimal.NewFromInt(in)).
			Add(NormalizePerM(rates.OutputPerM).Mul(decimal.NewFromInt(out))),
		Version: rates.PricingVersion,
	}

	// Cached read: apply the kernel hit ratio to the cached component.
	cachedReadPrice := NormalizePerM(rates.CachedReadPerM)
	b.CachedReadCost = cachedReadPrice.
		Mul(decimal.NewFromFloat(CacheHitRatio)).
		Mul(decimal.NewFromInt(cachedRead))

	if rates.CachedWritePerM != nil {
		b.CachedWriteCost = NormalizePerM(*rates.CachedWritePerM).Mul(decimal.NewFromInt(cachedWrite))
	}

	b.TotalCost = b.InputCost.Add(b.OutputCost).Add(b.CachedReadCost).Add(b.CachedWriteCost)
	return b, nil
}

// ComputeTokensPerDollar returns the display tokens per dollar for a
// blended per-token cost: 1M ÷ blend (kernel §4).
func ComputeTokensPerDollar(blendPerToken decimal.Decimal) decimal.Decimal {
	if !blendPerToken.IsPositive() {
		return decimal.Zero
	}
	return perMillion.Div(blendPerToken)
}

// ComputePercentPerMillion returns pct_per_1m = 100 ÷ tokens_per_dollar.
func ComputePercentPerMillion(tokensPerDollar decimal.Decimal) decimal.Decimal {
	if !tokensPerDollar.IsPositive() {
		return decimal.Zero
	}
	return decimal.NewFromInt(100).Div(tokensPerDollar)
}
