package service

import (
	"context"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

// supplierPricingCacheTTL 供应商定价缓存 TTL，与渠道缓存风格一致。
const supplierPricingCacheTTL = 5 * time.Minute

// SupplierPricingResolver 供应商定价解析器（带缓存 + singleflight）。
// 仅供内部路由成本排序使用。
type SupplierPricingResolver struct {
	repo  SupplierPricingRepository
	cache *gocache.Cache
	sf    singleflight.Group
}

func NewSupplierPricingResolver(repo SupplierPricingRepository) *SupplierPricingResolver {
	return &SupplierPricingResolver{
		repo:  repo,
		cache: gocache.New(supplierPricingCacheTTL, 2*supplierPricingCacheTTL),
	}
}

// ResolveActiveByModel 返回模型在 at 时刻的生效供应商定价，
// 结果按小写 supplier_code 规范化。空结果与错误不做长期缓存。
func (r *SupplierPricingResolver) ResolveActiveByModel(ctx context.Context, modelID string, at time.Time) ([]SupplierPricing, error) {
	cacheKey := modelID
	if v, ok := r.cache.Get(cacheKey); ok {
		return v.([]SupplierPricing), nil
	}
	v, err, _ := r.sf.Do(cacheKey, func() (any, error) {
		rows, err := r.repo.ListActiveByModel(ctx, modelID, at)
		if err != nil {
			return nil, err
		}
		normalized := make([]SupplierPricing, 0, len(rows))
		for _, row := range rows {
			row.SupplierCode = strings.ToLower(strings.TrimSpace(row.SupplierCode))
			normalized = append(normalized, row)
		}
		r.cache.Set(cacheKey, normalized, supplierPricingCacheTTL)
		return normalized, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]SupplierPricing), nil
}

// resolveSupplierPricingForModel 供 GatewayService 在选择循环外调用：
// 单模型一次解析（缓存+singleflight），解析失败时返回 nil（回退既有行为）。
func (s *GatewayService) resolveSupplierPricingForModel(ctx context.Context, modelID string) ([]SupplierPricing, error) {
	if s.supplierPricingResolver == nil || modelID == "" {
		return nil, nil
	}
	return s.supplierPricingResolver.ResolveActiveByModel(ctx, modelID, time.Now())
}

// rankBySupplierCost 计算账号在该模型下的供应商成本排名。
// 无有效定价或无法识别的供应商编码 → hasPricing=false。
func (s *GatewayService) rankBySupplierCost(pricing []SupplierPricing, account *Account) accountSelectionRank {
	code := account.SupplierCode()
	if code == "" {
		return accountSelectionRank{}
	}
	for _, p := range pricing {
		if p.SupplierCode != code {
			continue
		}
		if p.InputRate.IsNegative() || p.OutputRate.IsNegative() {
			return accountSelectionRank{}
		}
		if p.InputRate.IsZero() && p.OutputRate.IsZero() {
			return accountSelectionRank{}
		}
		return accountSelectionRank{
			hasPricing: true,
			cost:       p.InputRate.Add(p.OutputRate),
		}
	}
	return accountSelectionRank{}
}
