// Package zcloud contains the reseller platform extensions for zcloud/zrouter.
//
// This is the fork boundary: all zcloud-specific code lives under
// backend/zcloud/ so that upstream (Wei-Shaw/sub2api) changes can be
// rebased with minimal conflicts. Nothing in this package is required for
// the base sub2api gateway to run.
package zcloud

// FeatureFlags gates zcloud platform capabilities. Non-MVP sub2api features
// are disabled via config (never deleted), per kernel decision D22.
type FeatureFlags struct {
	// ModelPlans enables per-model token plans (1 plan = 1 model) with
	// stacked, no-expiry balances (kernel D6/D7/D15).
	ModelPlans bool `json:"model_plans" yaml:"model_plans"`

	// UsdtWallet enables TRON/TRC20 USDT wallet scanning and crediting
	// (kernel D5). Wallet watcher runs as a separate worker.
	UsdtWallet bool `json:"usdt_wallet" yaml:"usdt_wallet"`

	// ModelBalanceBlocks enables per-model 402 usage_cap_exhausted and
	// 403 model_unavailable rejection before routing (kernel D7).
	ModelBalanceBlocks bool `json:"model_balance_blocks" yaml:"model_balance_blocks"`

	// DualMetering enables per-model usage percent metering and
	// snapshots (kernel D8).
	DualMetering bool `json:"dual_metering" yaml:"dual_metering"`

	// ZcloudCustomerDashboard enables the customer-facing usage,
	// balances and export/delete pages (kernel D23).
	ZcloudCustomerDashboard bool `json:"zcloud_customer_dashboard" yaml:"zcloud_customer_dashboard"`

	// ZcloudAdmin enables admin model catalog and plan management
	// (kernel D20: single owner admin).
	ZcloudAdmin bool `json:"zcloud_admin" yaml:"zcloud_admin"`

	// Non-MVP sub2api features (kernel D22: disabled via config, not deleted).
	RedeemCodes   bool `json:"redeem_codes" yaml:"redeem_codes"`
	PromoCodes    bool `json:"promo_codes" yaml:"promo_codes"`
	Affiliate     bool `json:"affiliate" yaml:"affiliate"`
	Announcements bool `json:"announcements" yaml:"announcements"`
	BatchImage    bool `json:"batch_image" yaml:"batch_image"`
}

// DefaultFlags returns the production defaults for zcloud.
// Redeem/promo/affiliate/announcement/batch_image are OFF (D22).
func DefaultFlags() FeatureFlags {
	return FeatureFlags{
		ModelPlans:              true,
		UsdtWallet:              true,
		ModelBalanceBlocks:      true,
		DualMetering:            true,
		ZcloudCustomerDashboard: true,
		ZcloudAdmin:             true,
		RedeemCodes:             false,
		PromoCodes:              false,
		Affiliate:               false,
		Announcements:           false,
		BatchImage:              false,
	}
}
