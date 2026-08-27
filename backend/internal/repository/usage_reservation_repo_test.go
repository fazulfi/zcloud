package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/zcloud/billing"
	"github.com/shopspring/decimal"
)

func TestUsageReservationRepositoryReserve(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := &usageReservationRepository{db: db}
	req := billing.ReservationRequest{RequestID: "req-1", UserID: 7, AccountID: 9, ModelID: "m", Model: "model", Fingerprint: "fp", Cost: decimal.NewFromInt(2), PricingVersion: 1}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO usage_reservations")).WithArgs("req-1", int64(7), int64(9), "m", "model", "fp", req.Cost, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	got, err := r.Reserve(context.Background(), req)
	if err != nil || got.Status != billing.ReservationPending {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUsageReservationRepositoryConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := &usageReservationRepository{db: db}
	req := billing.ReservationRequest{RequestID: "req-1", Fingerprint: "new"}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO usage_reservations")).WithArgs("req-1", int64(0), int64(0), "", "", "new", req.Cost, 0).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, fingerprint FROM usage_reservations")).WithArgs("req-1").WillReturnRows(sqlmock.NewRows([]string{"status", "fingerprint"}).AddRow("pending", "old"))
	got, err := r.Reserve(context.Background(), req)
	if got.Status != "" || err != billing.ErrReservationConflict {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUsageReservationRepositoryNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := &usageReservationRepository{db: db}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE usage_reservations")).WithArgs("missing", "finalized").WillReturnResult(sqlmock.NewResult(0, 0))
	_, err = r.Finalize(context.Background(), billing.ReservationRequest{RequestID: "missing"})
	if err != billing.ErrReservationNotFound {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
