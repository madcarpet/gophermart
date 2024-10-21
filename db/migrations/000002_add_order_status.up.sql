BEGIN;

CREATE TABLE IF NOT EXISTS order_status (
    status_id UUID PRIMARY KEY,
    status_name VARCHAR(255) NOT NULL UNIQUE
);

INSERT INTO order_status (status_id, status_name) 
VALUES ('a1dd6e49-fa1d-42c4-8942-dfd675c7ba12', 'NEW');

INSERT INTO order_status (status_id, status_name) 
VALUES ('41969955-6807-4da8-beed-4b9b2681158e', 'PROCESSING');

INSERT INTO order_status (status_id, status_name) 
VALUES ('65fdd678-d9a2-4eb3-9b87-8f5e773cb4cf', 'INVALID');

INSERT INTO order_status (status_id, status_name) 
VALUES ('664c4011-d8bd-4576-957b-af054a2e5d0d', 'PROCESSED');

COMMIT;