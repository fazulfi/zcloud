package service

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Supplier 编码常量（内部供应商路由用）。
// 供应商编码是账号级元数据（Account.Extra["supplier_code"]），
// 不是上游平台（platform），也不是渠道（channel）标识。
const (
	SupplierCodeCB   = "cb"   // CodeBuddy
	SupplierCodeCBCN = "cbcn" // CodeBuddy CN
	SupplierCodeCX   = "cx"   // 其他/预留
)

// SupplierPricing 供应商定价 DTO（服务层表示，不暴露 ent 实体）。
// 数据来自 supplier_pricing 表，仅供内部路由与成本排序使用，
// 绝不对客户暴露。
type SupplierPricing struct {
	ModelID         string
	SupplierCode    string
	Version         int
	TierLabel       string
	Availability    string
	InputRate       decimal.Decimal
	OutputRate      decimal.Decimal
	CachedReadRate  decimal.Decimal
	CachedWriteRate decimal.Decimal
	EffectiveFrom   *time.Time
	EffectiveTo     *time.Time
}

// SupplierCode 返回账号的供应商编码（规范化后的小写形式）。
// 规则：
//   - 仅接受 Extra["supplier_code"] 中的字符串值；
//   - 去除首尾空白并转小写；
//   - 缺失、非字符串、或无法识别的值一律返回 ""（视为无定价/未标注供应商）；
//   - 绝不从账号 Name 推断供应商。
func (a *Account) SupplierCode() string {
	if a == nil || a.Extra == nil {
		return ""
	}
	raw, ok := a.Extra["supplier_code"]
	if !ok {
		return ""
	}
	s, ok := raw.(string)
	if !ok {
		return ""
	}
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case SupplierCodeCB, SupplierCodeCBCN, SupplierCodeCX:
		return s
	default:
		return ""
	}
}

// accountSelectionRank 账号在供应商成本排序中的排名信息。
type accountSelectionRank struct {
	// hasPricing 表示该账号存在该模型的有效供应商定价。
	hasPricing bool
	// cost 为加权成本（input_rate + output_rate，权重 1:1），
	// 仅当 hasPricing 为 true 时有效。
	cost decimal.Decimal
}

// isPreferredAccount 判断 candidate 是否比 selected 更优。
// 优先级（仅在全部 eligibility 过滤通过后调用）：
//  1. 有定价的账号优于无定价的账号；
//  2. 加权成本更低者优先；
//  3. 回退到既有调度比较器：Priority 升序 → 从未使用优先 →
//     preferOAuth（仅 gemini 平台）→ LastUsedAt 更早。
func (s *GatewayService) isPreferredAccount(candidate, selected *Account, candidateRank, selectedRank accountSelectionRank, preferOAuth bool) bool {
	// 1. 定价可用性分层：priced > unpriced
	if candidateRank.hasPricing != selectedRank.hasPricing {
		return candidateRank.hasPricing
	}
	if candidateRank.hasPricing {
		// 2. 加权成本更低者优先
		if cmp := candidateRank.cost.Cmp(selectedRank.cost); cmp != 0 {
			return cmp < 0
		}
	}
	// 3. 既有调度比较器
	if candidate.Priority != selected.Priority {
		return candidate.Priority < selected.Priority
	}
	// 从未使用优先
	cNeverUsed := candidate.LastUsedAt == nil
	sNeverUsed := selected.LastUsedAt == nil
	if cNeverUsed != sNeverUsed {
		return cNeverUsed
	}
	if preferOAuth {
		cOAuth := candidate.IsOAuth()
		sOAuth := selected.IsOAuth()
		if cOAuth != sOAuth {
			return cOAuth
		}
	}
	// LastUsedAt 更早优先
	if candidate.LastUsedAt != nil && selected.LastUsedAt != nil {
		return candidate.LastUsedAt.Before(*selected.LastUsedAt)
	}
	return false
}
