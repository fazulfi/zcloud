package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/zcloud/billing"
)

type fakeModelBalanceRepository struct {
	balance *ModelBalance
	err     error
}

func (r *fakeModelBalanceRepository) GetByUserAndModel(context.Context, int64, string) (*ModelBalance, error) {
	return r.balance, r.err
}
func (r *fakeModelBalanceRepository) SetBlocked(context.Context, int64, string, bool) error {
	return nil
}
func (r *fakeModelBalanceRepository) ListByUser(context.Context, int64) ([]ModelBalance, error) {
	return nil, nil
}
func (r *fakeModelBalanceRepository) CreditTokens(context.Context, int64, string, int64) error {
	return nil
}

func TestCheckModelAdmission(t *testing.T) {
	cases := []struct {
		name   string
		status string
		code   string
	}{
		{name: "no row"},
		{name: "active", status: "active"},
		{name: "blocked", status: "blocked", code: billing.CapErrorTypeUsageCapExhausted},
		{name: "not purchased", status: "not_purchased", code: billing.CapErrorTypeModelUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &GatewayService{modelBalanceRepo: &fakeModelBalanceRepository{balance: nil}}
			if tc.status != "" {
				s.modelBalanceRepo = &fakeModelBalanceRepository{balance: &ModelBalance{Status: tc.status}}
			}
			got, err := s.CheckModelAdmission(context.Background(), 1, "model-a")
			if err != nil || (got == nil) != (tc.code == "") || got != nil && got.Code != tc.code {
				t.Fatalf("got=%+v err=%v", got, err)
			}
		})
	}
	got, err := (&GatewayService{}).CheckModelAdmission(context.Background(), 1, "model-a")
	if got != nil || err != nil {
		t.Fatalf("nil repo should be no-op: got=%v err=%v", got, err)
	}
	_, err = (&GatewayService{modelBalanceRepo: &fakeModelBalanceRepository{err: errors.New("db")}}).CheckModelAdmission(context.Background(), 1, "model-a")
	if err == nil {
		t.Fatal("expected repository error")
	}
}
