-- +goose Up
CREATE PUBLICATION dbz_publication
FOR TABLE users, subscriptions;

-- +goose Down
DROP PUBLICATION IF EXISTS dbz_publication;