-- Seed the default group's customer-facing subscription plans for every catalog model and price tier.
SELECT pg_advisory_xact_lock(246247);

WITH models(canonical_name, public_name) AS (
    VALUES
        ('gpt-5.6-luna', 'GPT-5.6 Luna'),
        ('deepseek-v4-flash', 'DeepSeek V4 Flash'),
        ('gpt-5.4-mini', 'GPT-5.4 Mini'),
        ('minimax-m3', 'MiniMax M3'),
        ('minimax-m2.7', 'MiniMax M2.7'),
        ('gpt-5.6-terra', 'GPT-5.6 Terra'),
        ('deepseek-v4-pro', 'DeepSeek V4 Pro'),
        ('gpt-5.4', 'GPT-5.4'),
        ('glm-5.2', 'GLM-5.2'),
        ('glm-5.3', 'GLM-5.3'),
        ('kimi-k2.6', 'Kimi K2.6'),
        ('kimi-k2.7', 'Kimi K2.7'),
        ('gemini-3.1-pro', 'Gemini 3.1 Pro'),
        ('gpt-5.6-sol', 'GPT-5.6 Sol'),
        ('kimi-k3', 'Kimi K3'),
        ('gpt-5.5', 'GPT-5.5'),
        ('gpt-5.3-codex', 'GPT-5.3 [CD]')
), tiers(price, sort_order) AS (
    VALUES
        (1, 1),
        (2, 2),
        (5, 3),
        (10, 4)
)
INSERT INTO subscription_plans (
    group_id,
    name,
    price,
    currency,
    validity_days,
    for_sale,
    product_name,
    model_id,
    sort_order
)
SELECT
    (SELECT id FROM groups WHERE name = 'default' AND deleted_at IS NULL LIMIT 1),
    models.public_name || ' $' || tiers.price,
    tiers.price,
    'USD',
    30,
    TRUE,
    models.canonical_name,
    (SELECT id FROM model_catalog WHERE canonical_name = models.canonical_name LIMIT 1),
    tiers.sort_order
FROM models
CROSS JOIN tiers
WHERE NOT EXISTS (
    SELECT 1
    FROM subscription_plans existing
    WHERE existing.group_id = (SELECT id FROM groups WHERE name = 'default' AND deleted_at IS NULL LIMIT 1)
      AND existing.name = models.public_name || ' $' || tiers.price
);
