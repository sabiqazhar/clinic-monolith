# AGENTS.md

## Quick Start

```bash
# Start infrastructure
docker-compose up -d

# Run migrations (requires .env loaded)
make pg-up   # PostgreSQL modules (patient, billing)
make my-up   # MySQL module (appointment)

# Generate SQL code
sqlc generate
```

## Architecture

- **Modules**: `internal/modules/{patient,billing,appointment}`
- **Patient + Billing**: PostgreSQL
- **Appointment**: MySQL
- **DI**: `goforj/wire` (see `internal/modules/*/provider.go`)

## Key Commands

| Task | Command |
|------|---------|
| PostgreSQL up | `make pg-up` |
| MySQL up | `make my-up` |
| Create PG migration | `make create-pg name=migration_name` |
| Create MySQL migration | `make create-my name=migration_name` |
| Generate queries | `sqlc generate` |

## Dependencies

- PostgreSQL (port 5432), MySQL (3306), Redis (6379), RabbitMQ (5672)
- Run `docker-compose up -d` before any DB operation
