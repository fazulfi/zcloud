//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type m111Seed struct {
	UserID   int64
	GroupID  int64
	APIKeyID int64
	APIKey   string
	ModelIDs map[string]string
	PlanIDs  map[string]int64
}

var m111Models = []string{
	"gpt-5.6-luna", "deepseek-v4-flash", "gpt-5.4-mini", "minimax-m3",
	"minimax-m2.7", "gpt-5.6-terra", "deepseek-v4-pro", "gpt-5.4", "glm-5.2",
	"glm-5.3", "kimi-k2.6", "kimi-k2.7", "gemini-3.1-pro", "gpt-5.6-sol",
	"kimi-k3", "gpt-5.5", "gpt-5.3-codex", "gpt-5.6",
}

// seedM111Fixture creates a repeatable database fixture. Every natural key is scoped
// by the test name, so a test can safely invoke it more than once.
func seedM111Fixture(t *testing.T) m111Seed {
	t.Helper()
	ctx := context.Background()
	db := integrationDB
	scope := uniqueTestValue(t, "m111")
	seed := m111Seed{ModelIDs: make(map[string]string), PlanIDs: make(map[string]int64)}

	err := db.QueryRowContext(ctx, `INSERT INTO users(email,password_hash,role,status,concurrency)
		VALUES ($1,'m111-test-hash','user','active',1000)
		ON CONFLICT(email) DO UPDATE SET status='active' RETURNING id`, scope+"@example.com").Scan(&seed.UserID)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO groups(name,status) VALUES ($1,'active')
		ON CONFLICT(name) DO UPDATE SET status='active' RETURNING id`, scope+"-group").Scan(&seed.GroupID)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO api_keys(user_id,key,name,group_id,status,model_scope_mode)
		VALUES ($1,$2,'M1.11 integration key',$3,'active','all')
		ON CONFLICT(key) DO UPDATE SET status='active' RETURNING id,key`, seed.UserID, "sk-m111-"+scope, seed.GroupID).Scan(&seed.APIKeyID, &seed.APIKey)
	require.NoError(t, err)

	for _, model := range m111Models {
		var id string
		err = db.QueryRowContext(ctx, `INSERT INTO model_catalog(canonical_name,public_name,context_window,source_suppliers)
			VALUES ($1,$1,1000000,'["cb","cbcn","cx"]')
			ON CONFLICT(canonical_name) DO UPDATE SET status='active' RETURNING id`, model).Scan(&id)
		require.NoError(t, err)
		seed.ModelIDs[model] = id
		_, err = db.ExecContext(ctx, `INSERT INTO model_pricing(model_id,version,input_rate,output_rate,tokens_per_dollar,pct_per_1m_tokens,effective_from,source_ref)
			VALUES ($1,1,1,1,1000000,1,NOW(),'M1.11 test') ON CONFLICT(model_id,version) DO NOTHING`, id)
		require.NoError(t, err)
	}
	for _, model := range []string{"glm-5.2", "glm-5.3", "kimi-k3", "minimax-m3"} {
		for _, supplier := range []string{"cb", "cbcn"} {
			input, output := 2.0, 3.0
			if supplier == "cbcn" {
				input, output = 1.0, 1.5
			}
			_, err = db.ExecContext(ctx, `INSERT INTO supplier_pricing(model_id,supplier_code,version,tier_label,availability,input_rate,output_rate,effective_from)
				VALUES ($1,$2,1,NULL,'active',$3,$4,NOW()) ON CONFLICT(model_id,supplier_code,version,tier_label) DO UPDATE SET input_rate=EXCLUDED.input_rate,output_rate=EXCLUDED.output_rate`, seed.ModelIDs[model], supplier, input, output)
			require.NoError(t, err)
		}
	}
	for _, price := range []string{"1.00", "2.00", "5.00", "10.00"} {
		var planID int64
		err = db.QueryRowContext(ctx, `INSERT INTO subscription_plans(group_id,name,price,product_name,for_sale)
			VALUES ($1,$2,$3,$2,true) ON CONFLICT DO NOTHING RETURNING id`, seed.GroupID, scope+"-plan-"+price, price).Scan(&planID)
		if err != nil {
			// PostgreSQL has no natural uniqueness on plans; recover the ID after a rerun.
			err = db.QueryRowContext(ctx, `SELECT id FROM subscription_plans WHERE group_id=$1 AND name=$2`, seed.GroupID, scope+"-plan-"+price).Scan(&planID)
		}
		seed.PlanIDs[price] = planID
		require.NoError(t, err)
	}
	for i, model := range m111Models {
		status := "active"
		purchased, consumed := int64(1000000), int64(i*1000)
		if i == 1 {
			purchased, consumed = 1000, 999
		}
		if i == 2 {
			purchased, consumed, status = 1000, 1000, "blocked"
		}
		_, err = db.ExecContext(ctx, `INSERT INTO model_balances(user_id,model_id,tokens_purchased,tokens_consumed,balance,usage_percent,status)
			VALUES ($1,$2,$3,$4,$3-$4,($4::numeric/$3::numeric)*100,$5)
			ON CONFLICT(user_id,model_id) DO UPDATE SET tokens_purchased=EXCLUDED.tokens_purchased,tokens_consumed=EXCLUDED.tokens_consumed,balance=EXCLUDED.balance,usage_percent=EXCLUDED.usage_percent,status=EXCLUDED.status`, seed.UserID, model, purchased, consumed, status)
		require.NoError(t, err)
	}
	return seed
}

func m111RequestID(t *testing.T, suffix string) string {
	return fmt.Sprintf("%s-%s", uniqueTestValue(t, "m111-request"), suffix)
}
