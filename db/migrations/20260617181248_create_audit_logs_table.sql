-- +goose Up
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS audit_logs(
    id UUID PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,
    operation  VARCHAR(50) NOT NULL,
    before JSONB NOT NULL,
    after JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS audit_logs;
