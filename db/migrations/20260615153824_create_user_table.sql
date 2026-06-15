-- +goose Up
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,

    created_by UUID,
    updated_by UUID,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS users;
