//go:build integration

package repository

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestM111SeedIsIdempotentAndCoversCatalog(t *testing.T) {
	seed := seedM111Fixture(t)
	again := seedM111Fixture(t)
	require.Equal(t, seed.UserID, again.UserID)
	require.Len(t, seed.ModelIDs, 18)
	require.Len(t, seed.PlanIDs, 4)
	var modelCount, supplierCount, balanceCount int
	err := integrationDB.QueryRow(`SELECT count(*) FROM model_catalog WHERE canonical_name = ANY($1)`, "{"+joinM111Models()+"}").Scan(&modelCount)
	require.NoError(t, err)
	err = integrationDB.QueryRow(`SELECT count(*) FROM supplier_pricing WHERE supplier_code IN ('cb','cbcn')`).Scan(&supplierCount)
	require.NoError(t, err)
	err = integrationDB.QueryRow(`SELECT count(*) FROM model_balances WHERE user_id=$1`, seed.UserID).Scan(&balanceCount)
	require.NoError(t, err)
	require.Equal(t, 18, modelCount)
	require.GreaterOrEqual(t, supplierCount, 8)
	require.Equal(t, 18, balanceCount)
}

func joinM111Models() string {
	out := ""
	for i, model := range m111Models {
		if i > 0 {
			out += ","
		}
		out += `"` + model + `"`
	}
	return out
}

func TestM111ConcurrentModelBalanceDebitNeverOverspends(t *testing.T) {
	seed := seedM111Fixture(t)
	modelID := seed.ModelIDs[m111Models[0]]
	const attempts = 100
	errCh := make(chan error, attempts)
	var wg sync.WaitGroup
	wg.Add(attempts)
	for i := 0; i < attempts; i++ {
		go func() {
			errCh <- debitM111(seed.UserID, modelID, 150)
			wg.Done()
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
	var purchased, consumed, balance int64
	err := integrationDB.QueryRow(`SELECT tokens_purchased,tokens_consumed,balance FROM model_balances WHERE user_id=$1 AND model_id=$2`, seed.UserID, modelID).Scan(&purchased, &consumed, &balance)
	require.NoError(t, err)
	require.LessOrEqual(t, consumed, purchased)
	require.Equal(t, purchased-consumed, balance)
}

func debitM111(userID int64, modelID string, amount int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := integrationDB.ExecContext(ctx, `UPDATE model_balances SET tokens_consumed=tokens_consumed+$1,balance=balance-$1,usage_percent=((tokens_consumed+$1)::numeric/tokens_purchased::numeric)*100,updated_at=NOW()
		WHERE user_id=$2 AND model_id=$3 AND status='active' AND balance >= $1`, amount, userID, modelID)
	return err
}

func TestM111DuplicateReservationIsReplaySafe(t *testing.T) {
	seed := seedM111Fixture(t)
	requestID := m111RequestID(t, "duplicate")
	args := []any{requestID, seed.UserID, seed.APIKeyID, m111Models[0], "fingerprint", 1.0, "finalized"}
	_, err := integrationDB.Exec(`INSERT INTO usage_reservations(request_id,user_id,api_key_id,model,fingerprint,reserved_cost,status) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(request_id) DO NOTHING`, args...)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`INSERT INTO usage_reservations(request_id,user_id,api_key_id,model,fingerprint,reserved_cost,status) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(request_id) DO NOTHING`, args...)
	require.NoError(t, err)
	var count int
	err = integrationDB.QueryRow(`SELECT count(*) FROM usage_reservations WHERE request_id=$1`, requestID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestM111AdmissionStatusesAndScopeData(t *testing.T) {
	seed := seedM111Fixture(t)
	var exhausted string
	err := integrationDB.QueryRow(`SELECT status FROM model_balances WHERE user_id=$1 AND model_id=$2`, seed.UserID, seed.ModelIDs[m111Models[2]]).Scan(&exhausted)
	require.NoError(t, err)
	require.Equal(t, "blocked", exhausted)
	_, err = integrationDB.Exec(`INSERT INTO api_key_model_scopes(api_key_id,model_id,enabled) VALUES ($1,$2,true) ON CONFLICT(api_key_id,model_id) DO UPDATE SET enabled=true`, seed.APIKeyID, seed.ModelIDs[m111Models[0]])
	require.NoError(t, err)
	var enabled bool
	err = integrationDB.QueryRow(`SELECT enabled FROM api_key_model_scopes WHERE api_key_id=$1 AND model_id=$2`, seed.APIKeyID, seed.ModelIDs[m111Models[0]]).Scan(&enabled)
	require.NoError(t, err)
	require.True(t, enabled)
}

func TestM111LocalSupplierMockSupportsSuccessStreamAndFailure(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/fail" {
			http.Error(w, `{"error":"upstream_error"}`, http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/stream" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\ndata: [DONE]\n\n"))
			return
		}
		_, _ = w.Write([]byte(`{"id":"m111","choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()
	for _, path := range []string{"/chat", "/stream", "/fail"} {
		resp, err := http.Post(server.URL+path, "application/json", nil)
		require.NoError(t, err)
		require.NotNil(t, resp)
		_ = resp.Body.Close()
	}
	require.Equal(t, 3, calls)
}

func TestM111AdmissionLatencySanity(t *testing.T) {
	seed := seedM111Fixture(t)
	modelID := seed.ModelIDs[m111Models[0]]
	start := time.Now()
	for i := 0; i < 100; i++ {
		require.NoError(t, debitM111(seed.UserID, modelID, 1))
	}
	require.Less(t, time.Since(start), 5*time.Second)
}
