package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestAccountSupplierCode(t *testing.T) {
	tests := []struct {
		name  string
		extra map[string]any
		want  string
	}{
		{name: "missing key", extra: nil, want: ""},
		{name: "empty extra", extra: map[string]any{}, want: ""},
		{name: "valid cb", extra: map[string]any{"supplier_code": "cb"}, want: "cb"},
		{name: "valid cbcn", extra: map[string]any{"supplier_code": "cbcn"}, want: "cbcn"},
		{name: "valid cx", extra: map[string]any{"supplier_code": "cx"}, want: "cx"},
		{name: "uppercase normalized", extra: map[string]any{"supplier_code": "CB"}, want: "cb"},
		{name: "whitespace trimmed", extra: map[string]any{"supplier_code": "  cb  "}, want: "cb"},
		{name: "unknown value", extra: map[string]any{"supplier_code": "zz"}, want: ""},
		{name: "non-string value", extra: map[string]any{"supplier_code": 42}, want: ""},
		{name: "name does not affect", extra: map[string]any{"supplier_code": "cb"}, want: "cb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Account{Name: "CodeBuddy Account", Extra: tt.extra}
			if got := a.SupplierCode(); got != tt.want {
				t.Fatalf("SupplierCode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func newRank(hasPricing bool, cost string) accountSelectionRank {
	return accountSelectionRank{hasPricing: hasPricing, cost: decimal.RequireFromString(cost)}
}

func accWith(id int64, priority int, lastUsed *time.Time, accType string) *Account {
	return &Account{ID: id, Priority: priority, LastUsedAt: lastUsed, Type: accType}
}

func TestIsPreferredAccount(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	t.Run("cheap beats expensive", func(t *testing.T) {
		cheap := accWith(1, 100, &now, "")
		expensive := accWith(2, 100, &earlier, "")
		got := (&GatewayService{}).isPreferredAccount(cheap, expensive,
			newRank(true, "10"), newRank(true, "20"), false)
		if !got {
			t.Fatal("cheaper priced account should win")
		}
	})

	t.Run("priced beats unpriced", func(t *testing.T) {
		priced := accWith(1, 200, &now, "")
		unpriced := accWith(2, 100, &earlier, "")
		got := (&GatewayService{}).isPreferredAccount(priced, unpriced,
			newRank(true, "50"), newRank(false, "0"), false)
		if !got {
			t.Fatal("priced account should beat unpriced regardless of priority")
		}
	})

	t.Run("all unpriced falls back to priority", func(t *testing.T) {
		lowPrio := accWith(1, 100, &now, "")
		highPrio := accWith(2, 200, &earlier, "")
		got := (&GatewayService{}).isPreferredAccount(lowPrio, highPrio,
			newRank(false, "0"), newRank(false, "0"), false)
		if !got {
			t.Fatal("lower priority unpriced account should win")
		}
	})

	t.Run("equal cost falls back to priority", func(t *testing.T) {
		lowPrio := accWith(1, 100, &now, "")
		highPrio := accWith(2, 200, &now, "")
		got := (&GatewayService{}).isPreferredAccount(lowPrio, highPrio,
			newRank(true, "10"), newRank(true, "10"), false)
		if !got {
			t.Fatal("equal cost should fall back to lower priority")
		}
	})

	t.Run("equal cost and priority falls back to never-used", func(t *testing.T) {
		neverUsed := accWith(1, 100, nil, "")
		used := accWith(2, 100, &now, "")
		got := (&GatewayService{}).isPreferredAccount(neverUsed, used,
			newRank(true, "10"), newRank(true, "10"), false)
		if !got {
			t.Fatal("never-used should win on tie")
		}
	})

	t.Run("equal cost and priority prefers OAuth for gemini", func(t *testing.T) {
		oauth := accWith(1, 100, &now, AccountTypeOAuth)
		apikey := accWith(2, 100, &now, AccountTypeAPIKey)
		got := (&GatewayService{}).isPreferredAccount(oauth, apikey,
			newRank(true, "10"), newRank(true, "10"), true)
		if !got {
			t.Fatal("OAuth account should win for gemini on tie")
		}
	})

	t.Run("equal cost and priority falls back to oldest last used", func(t *testing.T) {
		older := accWith(1, 100, &earlier, "")
		newer := accWith(2, 100, &now, "")
		got := (&GatewayService{}).isPreferredAccount(older, newer,
			newRank(true, "10"), newRank(true, "10"), false)
		if !got {
			t.Fatal("older last-used should win on full tie")
		}
	})

	t.Run("invalid supplier code is unpriced", func(t *testing.T) {
		a := &Account{Extra: map[string]any{"supplier_code": "zz"}}
		if got := a.SupplierCode(); got != "" {
			t.Fatalf("unknown supplier code %q should be empty", got)
		}
	})
}

func TestRankBySupplierCost(t *testing.T) {
	svc := &GatewayService{}
	pricing := []SupplierPricing{
		{SupplierCode: "cb", InputRate: decimal.RequireFromString("1.40"), OutputRate: decimal.RequireFromString("4.40")},
		{SupplierCode: "cbcn", InputRate: decimal.RequireFromString("1.00"), OutputRate: decimal.RequireFromString("3.00")},
	}

	t.Run("cbcn cheaper than cb", func(t *testing.T) {
		acc := &Account{Extra: map[string]any{"supplier_code": "cbcn"}}
		rank := svc.rankBySupplierCost(pricing, acc)
		if !rank.hasPricing {
			t.Fatal("cbcn should be priced")
		}
		if !rank.cost.Equal(decimal.RequireFromString("4.00")) {
			t.Fatalf("cost = %s, want 4.00", rank.cost.String())
		}
	})

	t.Run("unknown code unpriced", func(t *testing.T) {
		acc := &Account{Extra: map[string]any{"supplier_code": "zz"}}
		if rank := svc.rankBySupplierCost(pricing, acc); rank.hasPricing {
			t.Fatal("unknown code must be unpriced")
		}
	})

	t.Run("negative rates unpriced", func(t *testing.T) {
		bad := []SupplierPricing{{SupplierCode: "cb", InputRate: decimal.RequireFromString("-1"), OutputRate: decimal.RequireFromString("2")}}
		acc := &Account{Extra: map[string]any{"supplier_code": "cb"}}
		if rank := svc.rankBySupplierCost(bad, acc); rank.hasPricing {
			t.Fatal("negative rate must be unpriced")
		}
	})

	t.Run("zero rates unpriced", func(t *testing.T) {
		bad := []SupplierPricing{{SupplierCode: "cb", InputRate: decimal.Zero, OutputRate: decimal.Zero}}
		acc := &Account{Extra: map[string]any{"supplier_code": "cb"}}
		if rank := svc.rankBySupplierCost(bad, acc); rank.hasPricing {
			t.Fatal("zero rates must be unpriced")
		}
	})

	t.Run("missing supplier code unpriced", func(t *testing.T) {
		acc := &Account{Extra: map[string]any{}}
		if rank := svc.rankBySupplierCost(pricing, acc); rank.hasPricing {
			t.Fatal("account without supplier code must be unpriced")
		}
	})
}

type fakeSupplierPricingRepo struct {
	rows []SupplierPricing
	err  error
}

func (f *fakeSupplierPricingRepo) ListActiveByModel(ctx context.Context, modelID string, at time.Time) ([]SupplierPricing, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]SupplierPricing, 0, len(f.rows))
	for _, r := range f.rows {
		if r.ModelID == modelID {
			out = append(out, r)
		}
	}
	return out, nil
}

func TestSupplierPricingResolver(t *testing.T) {
	repo := &fakeSupplierPricingRepo{rows: []SupplierPricing{
		{ModelID: "model-1", SupplierCode: "CB", InputRate: decimal.RequireFromString("1")}, // uppercase to test normalization
		{ModelID: "model-1", SupplierCode: "cx", InputRate: decimal.RequireFromString("2")},
	}}
	r := NewSupplierPricingResolver(repo)

	got, err := r.ResolveActiveByModel(context.Background(), "model-1", time.Now())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].SupplierCode != "cb" {
		t.Fatalf("supplier code not normalized: %q", got[0].SupplierCode)
	}

	if got2, _ := r.ResolveActiveByModel(context.Background(), "model-1", time.Now()); len(got2) != 2 {
		t.Fatal("cache hit should return rows")
	}

	if got3, _ := r.ResolveActiveByModel(context.Background(), "model-2", time.Now()); len(got3) != 0 {
		t.Fatal("separate model should not share cache entries")
	}
}
