package billing

import "testing"

func TestCapExhaustionErrorFor(t *testing.T) {
	blocked := CapExhaustionErrorFor("m1", ModelCapBlocked)
	if blocked.Code != CapErrorTypeUsageCapExhausted || blocked.Message != "model usage cap exhausted" || blocked.IsRetryable() {
		t.Fatalf("unexpected blocked error: %+v", blocked)
	}
	unavailable := CapExhaustionErrorFor("m2", ModelCapNotPurchased)
	if unavailable.Code != CapErrorTypeModelUnavailable || unavailable.Message != "model not available without a plan" {
		t.Fatalf("unexpected unavailable error: %+v", unavailable)
	}
	var err error = blocked
	if err.Error() == "" {
		t.Fatal("expected error message")
	}
}
