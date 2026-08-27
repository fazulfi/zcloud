package pricing

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCheapestActiveRate(t *testing.T) {
	now := time.Now()
	got, err := CheapestActiveRate([]CostRates{
		{SupplierCode: "expensive", InputPerM: decimal.NewFromInt(2), OutputPerM: decimal.NewFromInt(2), EffectiveFrom: now.Add(-time.Hour)},
		{SupplierCode: "cheap", InputPerM: decimal.NewFromInt(1), OutputPerM: decimal.NewFromInt(1), EffectiveFrom: now.Add(-time.Hour)},
	}, now)
	if err != nil || got.SupplierCode != "cheap" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestActiveCostRateWindow(t *testing.T) {
	now := time.Now()
	to := now.Add(time.Hour)
	if !ActiveCostRate(CostRates{EffectiveFrom: now.Add(-time.Hour), EffectiveTo: &to, Availability: "active"}, now) {
		t.Fatal("expected active rate")
	}
	if ActiveCostRate(CostRates{EffectiveFrom: now.Add(time.Hour), Availability: "active"}, now) {
		t.Fatal("future rate is active")
	}
}

func TestComputeCostUnknownPrice(t *testing.T) {
	_, err := ComputeCost(1, 1, 0, 0, CostRates{SupplierCode: "x"})
	if err != ErrUnknownPrice {
		t.Fatalf("err=%v", err)
	}
}
