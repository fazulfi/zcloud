package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelBalanceRepositoryNilClient(t *testing.T) {
	r := &modelBalanceRepository{}
	got, err := r.GetByUserAndModel(context.Background(), 1, "model")
	require.Error(t, err)
	require.Nil(t, got)
	require.Error(t, r.SetBlocked(context.Background(), 1, "model", true))
}
