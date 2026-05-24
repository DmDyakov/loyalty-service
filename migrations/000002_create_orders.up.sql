CREATE TABLE IF NOT EXISTS orders (
  id SERIAL PRIMARY KEY,
  number VARCHAR(255) NOT NULL UNIQUE,
  user_id INTEGER NOT NULL REFERENCES users(id),
  status VARCHAR(20) NOT NULL DEFAULT 'NEW'
    CHECK (status in ('NEW', 'INVALID', 'PROCESSING', 'PROCESSED')),
  accrual DECIMAL(10,2) DEFAULT 0 CHECK (accrual >= 0),
  uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
