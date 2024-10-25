CREATE TABLE IF NOT EXISTS orders_delayed (
    number VARCHAR(20) PRIMARY KEY,
    user_id UUID,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);