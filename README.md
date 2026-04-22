# clinic-monolith

Modular monolith example for a clinic management system in Go.

This project demonstrates a **bounded-context modular monolith** with:

- HTTP API via Gin
- PostgreSQL + MySQL split by module
- Redis cache
- RabbitMQ messaging
- SQLC-generated query layer
- Compile-time dependency injection with `goforj/wire`

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Run the Application](#run-the-application)
- [API Endpoints](#api-endpoints)
- [Development Workflow](#development-workflow)
- [Database & SQLC Workflow](#database--sqlc-workflow)
- [How the System Works (Request Flow)](#how-the-system-works-request-flow)
- [Observability](#observability)
- [Useful Commands](#useful-commands)
- [Notes](#notes)

---

## Overview

`clinic-monolith` organizes business capabilities into independent modules while keeping deployment as a single app.

Current modules:

- `patient` (PostgreSQL)
- `billing` (PostgreSQL)
- `appointment` (MySQL)

---

## Architecture

### Module design

Each module follows layered boundaries:

- `domain/` → contracts & domain errors/interfaces
- `repository/` → data access implementation
- `service/` → business logic orchestration
- `handler/` → HTTP transport layer
- `provider.go` → Wire provider set

### Infrastructure

- **PostgreSQL** for patient + billing domains
- **MySQL** for appointment domain
- **Redis** for cache abstraction used by services
- **RabbitMQ** for async event publishing
- **Outbox relay workers** for transactional event publication

### Dependency Injection

- Compile-time DI with `goforj/wire`
- Injector entry: `cmd/api/wire.go`
- Generated graph: `cmd/api/wire_gen.go`

---

## Project Structure

```text
cmd/api/
  main.go                 # app bootstrap, infra init, route mount, graceful shutdown
  wire.go                 # DI injector declarations
  wire_gen.go             # generated DI wiring

internal/
  config/                 # environment config loader
  infrastructure/
    db/                   # postgres/mysql constructors
    cache/                # redis client adapter
    broker/               # rabbitmq + outbox relay
  modules/
    patient/
    billing/
    appointment/

contracts/events/v1/      # shared event contracts
migrations/
  postgres/
  mysql/
sqlc.yaml                 # sqlc generation config (per module)
Makefile                  # migration helpers + module scaffold
docker-compose.yaml       # local infrastructure
```

---

## Prerequisites

- Go **1.25.8**
- Docker + Docker Compose
- [`migrate`](https://github.com/golang-migrate/migrate) CLI (required by `Makefile`)
- [`sqlc`](https://sqlc.dev/) CLI

---

## Setup

### 1) Clone and enter repository

```bash
git clone <your-repo-url>
cd clinic-monolith
```

### 2) Configure environment variables

```bash
cp .env.example .env
```

Adjust values in `.env` as needed.

Important keys used by app startup:

- `PG_URL`
- `MYSQL_URL_APP`
- `RABBITMQ_URL`
- `REDIS_HOST`, `REDIS_PORT`
- `SERVER` (optional, defaults to `8080`)

### 3) Start infrastructure

```bash
docker-compose up -d
```

Services started locally:

- PostgreSQL (`5432`)
- MySQL (`3306`)
- Redis (`6379`)
- RabbitMQ (`5672`, management UI `15672`)

### 4) Run database migrations

```bash
make pg-up
make my-up
```

### 5) Generate SQL code

```bash
sqlc generate
```

---

## Run the Application

```bash
go run ./cmd/api
```

Default app URL:

- `http://localhost:8080`

Health check:

- `GET http://localhost:8080/healthz`

pprof endpoint:

- `http://localhost:6060`

---

## API Endpoints

Base path: `/api/v1`

### Patient module

- `GET /api/v1/patients/:id`
- `POST /api/v1/patients/`

### Billing module

- `GET /api/v1/billing/:id`
- `POST /api/v1/billing/`

### Appointment module

Handlers exist and support:

- `GET /api/v1/appointments/:id`
- `POST /api/v1/appointments/`
- `DELETE /api/v1/appointments/:id`

---

## Development Workflow

Typical local workflow:

1. Start infra:

   ```bash
   docker-compose up -d
   ```

2. Apply migrations:

   ```bash
   make pg-up
   make my-up
   ```

3. Regenerate SQL layer when queries/schema change:

   ```bash
   sqlc generate
   ```

4. Run app:

   ```bash
   go run ./cmd/api
   ```

5. Test endpoints and inspect logs.

When changing DB schema:

- Create migration file(s)
- Apply migration(s)
- Update SQL queries if needed
- Re-run `sqlc generate`

---

## Database & SQLC Workflow

`sqlc.yaml` defines 3 separate generation targets:

- Patient (Postgres)
- Billing (Postgres)
- Appointment (MySQL)

This keeps generated query packages isolated per module.

Schema migration directories:

- `migrations/postgres`
- `migrations/mysql`

---

## How the System Works (Request Flow)

For a typical API request:

1. Gin router receives HTTP request.
2. Module `handler` validates payload and delegates to service.
3. Module `service` executes business logic.
4. Service calls `repository` for DB operations.
5. Service may read/write cache through cache abstraction.
6. Service may enqueue event via publisher abstraction.
7. Outbox relay background worker publishes queued events to RabbitMQ.

At startup (`cmd/api/main.go`), the app also:

- Loads `.env`
- Initializes logger
- Connects Postgres/MySQL/Redis/RabbitMQ
- Builds app dependencies via Wire injector
- Starts HTTP server + background relays
- Handles graceful shutdown on SIGINT/SIGTERM

---

## Observability

- Health endpoint: `/healthz`
- pprof server: `:6060`
- Structured logging via Zap

---

## Useful Commands

From `Makefile`:

```bash
# Create migration files
make create-pg name=<migration_name>
make create-my name=<migration_name>

# Apply rollback/check versions
make pg-down
make pg-force version=<N>
make pg-version

make my-down
make my-force version=<N>
make my-version

# Scaffold a new module skeleton
make create-module name=<module_name>
```

---

## Notes

- Ensure infrastructure is up before running migrations.
- Ensure migrations are applied before starting app.
- After any SQL/schema update, run `sqlc generate` to keep repository code in sync.
