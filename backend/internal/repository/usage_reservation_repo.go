package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/zcloud/billing"
)

type usageReservationRepository struct {
	db *sql.DB
}

// NewUsageReservationRepository creates the SQL-backed reservation repository.
func NewUsageReservationRepository(_ *dbent.Client, sqlDB *sql.DB) billing.ReservationRepository {
	return &usageReservationRepository{db: sqlDB}
}

func (r *usageReservationRepository) Reserve(ctx context.Context, req billing.ReservationRequest) (billing.ReservationResult, error) {
	if r == nil || r.db == nil {
		return billing.ReservationResult{}, errors.New("usage reservation repository db is nil")
	}
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO usage_reservations
			(request_id, user_id, api_key_id, account_id, model_id, model, fingerprint, reserved_cost, status, pricing_version)
		VALUES ($1, $2, 0, $3, $4, $5, $6, $7, 'pending', $8)
		ON CONFLICT (request_id) DO NOTHING
		RETURNING id
	`, strings.TrimSpace(req.RequestID), req.UserID, req.AccountID, req.ModelID, req.Model, req.Fingerprint, req.Cost, req.PricingVersion).Scan(&id)
	if err == nil {
		return billing.ReservationResult{Status: billing.ReservationPending}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return billing.ReservationResult{}, err
	}

	var status, fingerprint string
	if err := r.db.QueryRowContext(ctx, `
		SELECT status, fingerprint FROM usage_reservations WHERE request_id = $1
	`, strings.TrimSpace(req.RequestID)).Scan(&status, &fingerprint); err != nil {
		return billing.ReservationResult{}, err
	}
	if strings.TrimSpace(fingerprint) != strings.TrimSpace(req.Fingerprint) {
		return billing.ReservationResult{}, billing.ErrReservationConflict
	}
	if status == string(billing.ReservationPending) || status == string(billing.ReservationFinalized) {
		return billing.ReservationResult{Status: billing.ReservationStatus(status)}, nil
	}
	if status != string(billing.ReservationReleased) {
		return billing.ReservationResult{}, billing.ErrReservationConflict
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE usage_reservations
		SET user_id = $2, api_key_id = 0, account_id = $3, model_id = $4, model = $5,
			reserved_cost = $6, status = 'pending', pricing_version = $7, updated_at = now()
		WHERE request_id = $1 AND fingerprint = $8 AND status = 'released'
	`, strings.TrimSpace(req.RequestID), req.UserID, req.AccountID, req.ModelID, req.Model, req.Cost, req.PricingVersion, req.Fingerprint)
	if err != nil {
		return billing.ReservationResult{}, err
	}
	if rows, err := result.RowsAffected(); err != nil {
		return billing.ReservationResult{}, err
	} else if rows == 0 {
		return billing.ReservationResult{}, billing.ErrReservationConflict
	}
	return billing.ReservationResult{Status: billing.ReservationPending}, nil
}

func (r *usageReservationRepository) Finalize(ctx context.Context, req billing.ReservationRequest) (billing.ReservationResult, error) {
	return r.updateStatus(ctx, req.RequestID, billing.ReservationFinalized)
}

func (r *usageReservationRepository) Release(ctx context.Context, req billing.ReservationRequest) (billing.ReservationResult, error) {
	return r.updateStatus(ctx, req.RequestID, billing.ReservationReleased)
}

func (r *usageReservationRepository) updateStatus(ctx context.Context, requestID string, status billing.ReservationStatus) (billing.ReservationResult, error) {
	if r == nil || r.db == nil {
		return billing.ReservationResult{}, errors.New("usage reservation repository db is nil")
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE usage_reservations SET status = $2, updated_at = now() WHERE request_id = $1
	`, strings.TrimSpace(requestID), string(status))
	if err != nil {
		return billing.ReservationResult{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return billing.ReservationResult{}, err
	}
	if rows == 0 {
		return billing.ReservationResult{}, billing.ErrReservationNotFound
	}
	return billing.ReservationResult{Status: status}, nil
}

var _ billing.ReservationRepository = (*usageReservationRepository)(nil)
