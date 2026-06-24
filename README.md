# CDC Audit Timeline Service

CDC Audit Timeline Service is a Go service that captures database changes from PostgreSQL and writes them into an audit timeline. It combines a REST API for managing users and subscriptions with a Kafka consumer that listens to Debezium change events and stores normalized audit records in `audit_logs`.

The project solves a common backend problem: keeping an append-only audit history of important table changes without adding audit-write logic to every API handler or service method. PostgreSQL emits changes, Debezium publishes them to Kafka, and the consumer records those events as audit logs.

## What It Does

- Exposes REST APIs for `users` and `subscriptions`.
- Requires an `X-Requestor-Id` header for protected API routes.
- Uses PostgreSQL as the source database.
- Uses Debezium to stream PostgreSQL CDC events into Kafka.
- Runs a separate Go consumer process that reads Kafka CDC messages.
- Stores CDC changes in the `audit_logs` table with table name, operation, before state, after state, and timestamp.
- Includes unit tests for domain services, HTTP handlers, and Kafka message processing.
- Includes PostgreSQL-backed integration tests for repositories and the users, subscriptions, and audit-log HTTP APIs.

## Architecture

```text
HTTP client
   |
   v
API service: cmd/api
   |
   v
PostgreSQL: users, subscriptions
   |
   v
Debezium Connect
   |
   v
Kafka topics: cdc.public.users, cdc.public.subscriptions
   |
   v
Consumer service: cmd/consumer
   |
   v
PostgreSQL: audit_logs
```

<img width="1010" height="840" alt="CDC Audit Timeline Service" src="https://github.com/user-attachments/assets/2054b48b-1f5e-4f01-a7ef-74e0a51ea6b8" />


## Folder Structure

```text
.
|-- cmd/
|   |-- api/                  # API executable entrypoint
|   `-- consumer/             # CDC consumer executable entrypoint
|-- db/
|   `-- migrations/           # Goose database migrations and seed data
|-- deployment/
|   `-- Dockerfile            # Builds API and consumer binaries
|-- internal/
|   |-- app/                  # Application wiring for API and consumer
|   |-- config/               # Typed environment configuration
|   |-- domain/               # Business models, services, and store ports
|   |-- errors/               # Shared app errors
|   |-- infrastructure/       # Postgres repositories and logger setup
|   |-- transport/            # HTTP and Kafka transport code
|   `-- validation/           # Request validation formatting
|-- scripts/
|   `-- register-connector.sh # Helper script to register Debezium connector
|-- docker-compose.yml
|-- makefile
`-- README.md
```

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- Goose CLI for migrations

Install Goose if needed:

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Environment Setup

Create local env files:

```sh
cp .env.example .env
cp postgres.example.env postgres.env
```

For local Go runs against Docker Postgres, `.env` should use the published host port:

```env
PORT=:8000
DB_USER=postgres
DB_PASSWORD=postgres
DB_HOST=127.0.0.1
DB_PORT=5434
DB_NAME=cdc_audit_timeline_service
DB_SSLMODE=disable
KAFKA_BROKERS=localhost:29092
KAFKA_TOPICS=cdc.public.users,cdc.public.subscriptions
KAFKA_GROUP_ID=cdc-audit-consumer
```

Docker Compose overrides container networking automatically:

```text
DB_HOST=postgres
DB_PORT=5432
KAFKA_BROKERS=kafka:9092
KAFKA_TOPICS=cdc.public.users,cdc.public.subscriptions
```

## Run With Docker Compose

Start all services:

```sh
docker compose up --build
```

This starts:

- `postgres`
- `kafka`
- `debezium`
- `debezium-connector`
- `api`
- `consumer`

The Compose file includes healthchecks so the API waits for PostgreSQL, and the consumer waits for PostgreSQL, Kafka, and the Debezium connector registration job.

Run migrations after PostgreSQL is healthy:

```sh
make migrate-up
```

The `debezium-connector` service registers the connector automatically after Debezium is healthy. You can rerun only the connector registration job if needed:

```sh
docker compose up debezium-connector
```

Check API health:

```sh
curl http://localhost:8000/api/health
```

Expected response:

```json
{"message":"OK"}
```

## Run Locally

Start infrastructure only:

```sh
docker compose up -d postgres kafka debezium
```

Run migrations:

```sh
make migrate-up
```

Register Debezium:

```sh
docker compose up debezium-connector
```

Run the API:

```sh
make run
```

Run the CDC consumer in a second terminal:

```sh
make run-consumer
```

## API Usage

Protected routes require this header:

```text
X-Requestor-Id: 11111111-1111-1111-1111-111111111111
```

That user is created by the seed migration.

List users:

```sh
curl -H "X-Requestor-Id: 11111111-1111-1111-1111-111111111111" \
  "http://localhost:8000/api/users"
```

Create a user:

```sh
curl -X POST http://localhost:8000/api/users/ \
  -H "Content-Type: application/json" \
  -H "X-Requestor-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{
    "first_name": "Maya",
    "last_name": "Kapoor",
    "email": "maya.kapoor@example.com"
  }'
```

Create a subscription:

```sh
curl -X POST http://localhost:8000/api/subscriptions/ \
  -H "Content-Type: application/json" \
  -H "X-Requestor-Id: 11111111-1111-1111-1111-111111111111" \
  -d '{
    "user_id": "11111111-1111-1111-1111-111111111111",
    "plan_name": "basic",
    "status": "active",
    "start_date": "2026-06-01T00:00:00Z",
    "end_date": "2026-07-01T00:00:00Z",
    "auto_renew": true
  }'
```

List audit logs:

```sh
curl -H "X-Requestor-Id: 11111111-1111-1111-1111-111111111111" \
  "http://localhost:8000/api/audit-logs/"
```

List audit logs with pagination and sorting:

```sh
curl -H "X-Requestor-Id: 11111111-1111-1111-1111-111111111111" \
  "http://localhost:8000/api/audit-logs/?limit=20&page=1&orderBy=created_at&sortBy=desc"
```

### Audit Log API

`GET /api/audit-logs/`

Returns audit entries recorded by the CDC consumer from Debezium events.

Supported query parameters:

- `limit`: number of rows per page. Default is `10`.
- `page`: page number starting from `1`. Default is `1`.
- `orderBy`: one of `id`, `table_name`, `operation`, or `created_at`. Default is `created_at`.
- `sortBy`: `asc` or `desc`. Default is `asc`.

Example response:

```json
{
  "audit_logs": [
    {
      "id": "e8ef7d32-5fd6-4c6d-a248-f5d0db98aa8c",
      "table_name": "users",
      "operation": "update",
      "before": {
        "id": "11111111-1111-1111-1111-111111111111",
        "first_name": "Maya",
        "last_name": "Kapoor",
        "email": "maya.old@example.com"
      },
      "after": {
        "id": "11111111-1111-1111-1111-111111111111",
        "first_name": "Maya",
        "last_name": "Kapoor",
        "email": "maya.new@example.com"
      },
      "changes": {
        "email": {
          "old": "maya.old@example.com",
          "new": "maya.new@example.com"
        }
      },
      "created_at": "2026-06-19T10:30:00Z"
    }
  ],
  "total_results": 1
}
```

`changes` includes only the fields whose values changed. Each entry contains the previous value in `old` and the updated value in `new`.

Possible validation errors:

- `400 Bad Request` with `{"error":"invalid limit"}`
- `400 Bad Request` with `{"error":"invalid page"}`
- `400 Bad Request` with `{"error":"orderBy must be one of id, table_name, operation, created_at"}`
- `400 Bad Request` with `{"error":"sortBy must be asc or desc"}`

## CDC Flow

Once migrations and the Debezium connector are active:

1. Create, update, or delete rows in watched tables.
2. PostgreSQL logical replication exposes the change.
3. Debezium publishes the change event to Kafka.
4. The Go consumer reads the Kafka event.
5. The consumer writes an audit record to `audit_logs`.

Current default consumer topics:

```text
cdc.public.users
cdc.public.subscriptions
```

To consume a different set of topics, set:

```env
KAFKA_TOPICS=cdc.public.users,cdc.public.subscriptions
```

## Debezium Connector

Docker Compose registers the Postgres CDC connector through the `debezium-connector` one-shot service. It uses `PUT` so the operation is idempotent: rerunning it creates the connector if missing or updates the connector config if it already exists.

Manual registration is still useful for troubleshooting:

```sh
curl -X PUT http://localhost:8083/connectors/postgres-cdc-connector/config \
  -H "Content-Type: application/json" \
  -d '{
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "debezium",
    "database.password": "debezium",
    "database.dbname": "cdc_audit_timeline_service",
    "topic.prefix": "cdc",
    "plugin.name": "pgoutput",
    "publication.name": "dbz_publication",
    "slot.name": "debezium_slot",
    "table.include.list": "public.users,public.subscriptions"
  }'
```

Check connector status:

```sh
curl http://localhost:8083/connectors/postgres-cdc-connector/status
```

## Validate Connectors

Use these links after the stack is running:

- List registered connectors: <http://localhost:8083/connectors>
- Check connector status: <http://localhost:8083/connectors/postgres-cdc-connector/status>
- View connector config: <http://localhost:8083/connectors/postgres-cdc-connector/config>
- View connector tasks: <http://localhost:8083/connectors/postgres-cdc-connector/tasks>

Expected connector status should include:

```json
{
  "connector": {
    "state": "RUNNING"
  },
  "tasks": [
    {
      "state": "RUNNING"
    }
  ]
}
```

Validate Kafka topics from the Kafka container:

```sh
docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```

Expected CDC topics:

```text
cdc.public.users
cdc.public.subscriptions
```

## Useful Commands

```sh
make up               # docker compose up -d
make rebuild          # docker compose up --build
make logs             # follow compose logs
make down             # stop compose services
make migrate-up       # apply DB migrations
make migrate-down     # roll back one migration
make run              # run API locally
make run-consumer     # run consumer locally
make test             # run tests
make integration-test # run integration tests
make fmt              # format Go files
make build            # build API and consumer binaries
```

## Testing

The project has two test suites:

- **Unit tests** run without PostgreSQL, Kafka, or Docker. They cover domain-service behaviour, HTTP handler validation and responses, and CDC consumer message handling.
- **Integration tests** use a dedicated PostgreSQL database and cover PostgreSQL repositories plus the users, subscriptions, and audit-log API flows end to end.

### Run Unit Tests

From the project root, run:

```sh
make test
```

This is equivalent to `go test ./...` and excludes tests marked with the `integration` build tag.

### Run Integration Tests

Integration tests use the isolated database configured in `postgres.test.env` (port `5435` by default); they do not use the development database.

1. Create the local test environment file if it does not already exist:

   ```sh
   cp postgres.test.example.env postgres.test.env
   ```

2. Start the test PostgreSQL container:

   ```sh
   docker compose -f docker-compose.test.yml up -d
   ```

3. Apply the migrations to the test database:

   ```sh
   make migrate-up-test
   ```

4. Run the tagged integration suite sequentially:

   ```sh
   make integration-test
   ```

`make integration-test` runs `go test -p 1 -tags integration ./...`. The `-p 1` setting prevents concurrently running packages from resetting the shared test database at the same time.

When finished, stop the test database with:

```sh
docker compose -f docker-compose.test.yml down
```

## Notes

- API and consumer are separate services because they have different runtime responsibilities.
- Domain packages define business models, services, and store interfaces.
- Infrastructure packages implement those store interfaces using PostgreSQL.
- Transport packages contain framework-specific code such as Gin handlers and Kafka message parsing.
- Debezium connector registration is automated by Compose, but the script remains available for manual troubleshooting.
