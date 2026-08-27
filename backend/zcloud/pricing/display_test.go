package pricing

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestComputeDisplayBlendMath(t *testing.T) {
	rates := DisplayRates{InputPerM: decimal.NewFromFloat(3), OutputPerM: decimal.NewFromFloat(15), CachedReadPerM: decimal.NewFromFloat(0.3), PricingVersion: PricingVersionDualMetering}
	got, err := ComputeDisplay(1_000_000, 100_000, 0, 0, rates)
	if err != nil {
		t.Fatal(err)
	}
	if !got.TotalCost.Equal(decimal.NewFromFloat(4.5)) {
		t.Fatalf("total cost = %s", got.TotalCost)
	}
	if !got.BlendCost.Equal(decimal.NewFromFloat(2.1795)) {
		t.Fatalf("blend cost = %s", got.BlendCost)
	}
}
