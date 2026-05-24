CREATE TABLE IF NOT EXISTS withdrawals (
  id SERIAL PRIMARY KEY,
  order_number VARCHAR(255) NOT NULL,
  user_id INTEGER NOT NULL REFERENCES users(id),
  sum DECIMAL(10,2) NOT NULL CHECK (sum > 0),
  processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);