-- +goose Up
GRANT CONNECT ON DATABASE cdc_audit_timeline_service TO debezium;
GRANT USAGE ON SCHEMA public TO debezium;

GRANT SELECT ON users TO debezium;
GRANT SELECT ON subscriptions TO debezium;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT ON TABLES TO debezium;

-- +goose Down
REVOKE SELECT ON users FROM debezium;
REVOKE SELECT ON subscriptions FROM debezium;
REVOKE USAGE ON SCHEMA public FROM debezium;
REVOKE CONNECT ON DATABASE cdc_audit_timeline_service FROM debezium;