// Package billing implements the atomic per-model balance reservation
// used by the dual metering pipeline (kernel §5).
//
// Reserve is performed once per request before dispatch; Finalize
// commits the actual meter on completion. Partial streams are settled
// by the reconciler which finalizes any pending reservation older than
// the staleness window (kernel §5 "reconciliation").
package billing

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// ReservationStatus describes the lifecycle of a model balance hold.
type ReservationStatus string

const (
	// ReservationPending marks a hold that is reserved but not yet settled.
	ReservationPending ReservationStatus = "pending"
	// ReservationFinalized marks a hold that was settled atomically.
	ReservationFinalized ReservationStatus = "finalized"
	// ReservationReleased marks a hold that was cancelled without usage.
	ReservationReleased ReservationStatus = "released"
)

// ErrReservationConflict is returned when the same request_id is
// reserved twice with a different fingerprint (idempotency key).
var ErrReservationConflict = errors.New("billing: reservation conflict")

// ErrReservationNotFound is returned when finalize/release references an
// unknown request_id.
var ErrReservationNotFound = errors.New("billing: reservation not found")

// ReservationRequest is the unit of atomic model balance hold.
//
// Cost is the display$ cost of the request (customer-visible). The
// supplier cost$ is stored separately in the meter and never exposed.
type ReservationRequest struct {
	RequestID      string
	UserID         int64
	ModelID        string // canonical model id
	Model          string // canonical model name
	Fingerprint    string
	Cost           decimal.Decimal
	AccountID      int64
	PricingAt      time.Time
	PricingVersion int
}

// ReservationResult is returned by Reserve/Finalize/Release.
type ReservationResult struct {
	Status     ReservationStatus
	NewBalance decimal.Decimal
	Overdrawn  bool
}

// ReservationRepository persists model balance holds.
//
// Implementations must be atomic per request_id: the idempotency key is
// (request_id, user_id) and the fingerprint guards against conflicts.
type ReservationRepository interface {
	// Reserve creates a pending hold for the request. Idempotent by
	// request_id: a repeated reserve with the same fingerprint returns
	// the existing hold; a different fingerprint returns
	// ErrReservationConflict.
	Reserve(ctx context.Context, req ReservationRequest) (ReservationResult, error)
	// Finalize settles a pending hold with the actual meter value.
	// Idempotent by request_id.
	Finalize(ctx context.Context, req ReservationRequest) (ReservationResult, error)
	// Release cancels a pending hold without usage. Idempotent.
	Release(ctx context.Context, req ReservationRequest) (ReservationResult, error)
}
