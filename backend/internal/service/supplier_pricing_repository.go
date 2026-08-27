package service

import (
	"context"
	"time"
)

// SupplierPricingRepository 供应商定价仓储接口。
// 数据仅供内部路由成本排序，绝不对客户暴露。
type SupplierPricingRepository interface {
	// ListActiveByModel 返回指定模型在 at 时刻处于生效期的供应商定价。
	// 过滤条件：availability='active' 且 effective_from<=at 且
	// (effective_to IS NULL OR effective_to>at)。无匹配行返回空切片。
	ListActiveByModel(ctx context.Context, modelID string, at time.Time) ([]SupplierPricing, error)
}
