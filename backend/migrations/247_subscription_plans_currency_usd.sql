-- Repair subscription plan display currencies so all saleable plans use USD.
SELECT pg_advisory_xact_lock(247247);

UPDATE subscription_plans
SET currency = 'USD'
WHERE currency IS DISTINCT FROM 'USD';
