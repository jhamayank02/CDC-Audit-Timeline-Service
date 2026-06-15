-- +goose Up
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),

    plan_name VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,

    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,

    auto_renew BOOLEAN DEFAULT TRUE,

    created_by UUID,
    updated_by UUID,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS subscriptions;