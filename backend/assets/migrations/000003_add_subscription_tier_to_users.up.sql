ALTER TABLE users ADD COLUMN subscription_tier TEXT NOT NULL DEFAULT 'free';
CREATE INDEX idx_users_subscription_tier ON users(subscription_tier);

