CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(20) PRIMARY KEY,
    status_id UUID,
    FOREIGN KEY (status_id) REFERENCES order_status(status_id) ON DELETE RESTRICT,
    user_id UUID,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    accrual real,
    upload_at TIMESTAMPTZ
);