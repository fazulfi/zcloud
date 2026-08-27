package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/zcloud/billing"
)

func (s *GatewayService) CheckModelAdmission(ctx context.Context, userID int64, model string) (*billing.CapExhaustionError, error) {
	if s == nil || s.modelBalanceRepo == nil {
		return nil, nil
	}
	balance, err := s.modelBalanceRepo.GetByUserAndModel(ctx, userID, model)
	if err != nil {
		return nil, err
	}
	if balance == nil {
		return nil, nil
	}
	switch billing.ModelCapStatus(balance.Status) {
	case billing.ModelCapBlocked:
		return billing.CapExhaustionErrorFor(model, billing.ModelCapBlocked), nil
	case billing.ModelCapNotPurchased:
		return billing.CapExhaustionErrorFor(model, billing.ModelCapNotPurchased), nil
	default:
		return nil, nil
	}
}
