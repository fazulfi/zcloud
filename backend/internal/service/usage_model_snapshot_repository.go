package service

import "context"

// UsageModelSnapshotRepository persists per-user per-model dual meter
// rollups (M1.7). Implementations must be idempotent per (user_id, model,
// pricing_version) row key.
type UsageModelSnapshotRepository interface {
	Upsert(ctx context.Context, snap *UsageModelSnapshot) error
}
