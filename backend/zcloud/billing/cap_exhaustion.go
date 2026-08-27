package billing

const (
	CapErrorTypeUsageCapExhausted = "usage_cap_exhausted"
	CapErrorTypeModelUnavailable  = "model_unavailable"
)

// ModelCapStatus describes the admission state of a user's model plan.
type ModelCapStatus string

const (
	ModelCapActive       ModelCapStatus = "active"
	ModelCapBlocked      ModelCapStatus = "blocked"
	ModelCapNotPurchased ModelCapStatus = "not_purchased"
)

// CapExhaustionError is the admission contract for model-specific billing gates.
type CapExhaustionError struct {
	Model   string
	Status  ModelCapStatus
	Code    string
	Message string
}

func (e *CapExhaustionError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// IsRetryable reports whether retrying the same model can change the admission result.
func (e *CapExhaustionError) IsRetryable() bool { return false }

// CapExhaustionErrorFor creates the stable admission error for a model cap state.
func CapExhaustionErrorFor(model string, status ModelCapStatus) *CapExhaustionError {
	err := &CapExhaustionError{Model: model, Status: status}
	switch status {
	case ModelCapBlocked:
		err.Code = CapErrorTypeUsageCapExhausted
		err.Message = "model usage cap exhausted"
	case ModelCapNotPurchased:
		err.Code = CapErrorTypeModelUnavailable
		err.Message = "model not available without a plan"
	}
	return err
}

var _ error = (*CapExhaustionError)(nil)
