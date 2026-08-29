package service

import (
	"context"
	"testing"
)

func TestGetLatestTokensPerDollarReturnsPricing(t *testing.T) {
	client := newPaymentConfigServiceTestClient(t)
	ctx := context.Background()
	model := client.ModelCatalog.Create().
		SetCanonicalName("token-test-model").
		SetPublicName("Token Test Model").
		SaveX(ctx)

	client.ModelPricing.Create().
		SetModelID(model.ID).
		SetVersion(1).
		SetTokensPerDollar(2400).
		SaveX(ctx)

	svc := &PaymentConfigService{entClient: client}
	got, err := svc.GetLatestTokensPerDollar(ctx, model.ID)
	if err != nil {
		t.Fatalf("GetLatestTokensPerDollar returned error: %v", err)
	}
	if got != 2400 {
		t.Fatalf("GetLatestTokensPerDollar = %d, want 2400", got)
	}
}
