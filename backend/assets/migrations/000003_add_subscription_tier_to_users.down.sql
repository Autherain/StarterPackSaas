DROP INDEX IF EXISTS idx_users_subscription_tier;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_tier;

