CREATE TABLE IF NOT EXISTS withdrawals (
    id UUID PRIMARY KEY,
    user_id UUID,
    order_num VARCHAR(20),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    summ real,
    processed_at TIMESTAMPTZ
);