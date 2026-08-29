package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func testGomerch(t *testing.T, handler http.Handler) *Gomerch {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	g, err := NewGomerch("test", map[string]string{"apiKey": "secret", "apiBase": ts.URL})
	require.NoError(t, err)
	return g
}

func TestGomerchCreatePayment(t *testing.T) {
	g := testGomerch(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/qris/create", r.URL.Path)
		require.Equal(t, "secret", r.Header.Get("X-API-Key"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"amount":1600000,"static_qr":false}`, string(body))
		_, err = w.Write([]byte(`{"success":true,"payment_ref":"550e8400-e29b-41d4-a716-446655440000","amount":1600000,"amount_unit":"major_idr","status":"pending","qr_string":"qr-data"}`))
		require.NoError(t, err)
	}))

	got, err := g.CreatePayment(context.Background(), payment.CreatePaymentRequest{Amount: "1600000"})
	require.NoError(t, err)
	require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", got.TradeNo)
	require.Equal(t, "qr-data", got.QRCode)
	require.Empty(t, got.PayURL)
	require.Equal(t, "IDR", got.Currency)
	require.Equal(t, payment.CreatePaymentResultOrderCreated, got.ResultType)
}

func TestGomerchQueryOrder(t *testing.T) {
	tests := []struct {
		name           string
		upstreamStatus string
		wantStatus     string
	}{
		{name: "paid", upstreamStatus: "PAID", wantStatus: payment.ProviderStatusPaid},
		{name: "pending", upstreamStatus: "PENDING", wantStatus: payment.ProviderStatusPending},
		{name: "expired", upstreamStatus: "EXPIRED", wantStatus: payment.ProviderStatusFailed},
		{name: "ambiguous", upstreamStatus: "AMBIGUOUS", wantStatus: payment.ProviderStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := testGomerch(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/api/qris/status", r.URL.Path)
				require.Equal(t, "secret", r.Header.Get("X-API-Key"))
				var body map[string]string
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", body["payment_ref"])
				_, err := w.Write([]byte(`{"success":true,"payment_ref":"550e8400-e29b-41d4-a716-446655440000","status":"` + tt.upstreamStatus + `","amount":1600000,"amount_unit":"major_idr"}`))
				require.NoError(t, err)
			}))

			got, err := g.QueryOrder(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
			require.NoError(t, err)
			require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", got.TradeNo)
			require.Equal(t, tt.wantStatus, got.Status)
			require.Equal(t, float64(1600000), got.Amount)
		})
	}
}

func TestGomerchVerifyNotification(t *testing.T) {
	g, err := NewGomerch("test", map[string]string{"apiKey": "secret"})
	require.NoError(t, err)
	got, err := g.VerifyNotification(context.Background(), `{}`, nil)
	require.Nil(t, got)
	require.EqualError(t, err, "gomerch does not support webhooks")
}

func TestGomerchRefundUnsupported(t *testing.T) {
	g, err := NewGomerch("test", map[string]string{"apiKey": "secret"})
	require.NoError(t, err)
	_, err = g.Refund(context.Background(), payment.RefundRequest{})
	require.ErrorIs(t, err, ErrRefundUnsupported)
}

func TestGomerchCancelPayment(t *testing.T) {
	g, err := NewGomerch("test", map[string]string{"apiKey": "secret"})
	require.NoError(t, err)
	require.NoError(t, g.CancelPayment(context.Background(), "550e8400-e29b-41d4-a716-446655440000"))
}
