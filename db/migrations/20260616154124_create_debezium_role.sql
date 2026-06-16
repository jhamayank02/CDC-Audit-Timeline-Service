-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_roles WHERE rolname = 'debezium'
  ) THEN
    CREATE ROLE debezium WITH LOGIN PASSWORD 'debezium' REPLICATION;
  END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
DROP ROLE IF EXISTS debezium;