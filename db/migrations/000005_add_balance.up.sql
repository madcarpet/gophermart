CREATE TABLE IF NOT EXISTS balance (
    id UUID PRIMARY KEY,
    user_id UUID,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    current real,
    withdrawn real
);