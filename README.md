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

## How to Create a New Module (Step-by-Step Guide for Interns)

Hey! 👋 So you need to add a new feature/module to this project? Don't worry—this guide will walk you through everything step by step. We'll use the example of creating a `doctor` module, but you can replace `doctor` with whatever module name you need.

### Overview: What We're Building

Each module in this project follows the same layered structure:

```
internal/modules/<module_name>/
├── domain/          → Defines data structures & interfaces (the "contracts")
├── repository/      → Database access layer (talks to SQL)
│   └── query/       → Auto-generated SQL code by sqlc
├── service/         → Business logic (orchestrates repository + cache + events)
├── handler/         → HTTP layer (receives requests, returns JSON responses)
└── provider.go      → Wire dependency injection configuration
```

The flow is: **HTTP Request → Handler → Service → Repository → Database**

---

### Step 1: Scaffold the Module Structure

First, run the Makefile command to create the basic folder structure:

```bash
make create-module name=doctor
```

This creates:
- `internal/modules/doctor/domain/interfaces.go`
- `internal/modules/doctor/handler/doctor.go`
- `internal/modules/doctor/repository/doctor.go`
- `internal/modules/doctor/repository/queries.sql`
- `internal/modules/doctor/service/doctor.go`
- `internal/modules/doctor/provider.go`

✅ **What you get:** Empty files with package declarations. Now let's fill them in!

---

### Step 2: Define Domain Layer (Contracts & Data Structures)

Open `internal/modules/doctor/domain/interfaces.go` and define:
- Your entity (data structure)
- Repository interface (what the repo must do)
- Service interface (what the service must do)
- Infrastructure interfaces (cache, event publisher—already defined, just reuse them)

Example:

```go
package domain

import (
	"context"
	"errors"
	"time"
)

// Domain Errors
var (
	ErrDoctorNotFound = errors.New("doctor not found")
)

// Entity
type Doctor struct {
	ID        string
	Name      string
	Specialty string
	Email     string
	CreatedAt time.Time
}

// Repository Interface
type DoctorRepository interface {
	FindByID(ctx context.Context, id string) (*Doctor, error)
	SaveWithOutbox(ctx context.Context, d *Doctor) error
}

// Service Interface (Public Contract)
type DoctorService interface {
	GetProfile(ctx context.Context, id string) (*Doctor, error)
	Register(ctx context.Context, name, specialty, email string) (*Doctor, error)
}
```

💡 **Tip:** Look at `patient/domain/interfaces.go` for reference. Copy the pattern!

---

### Step 3: Create Database Migration

Before writing queries, you need a database table. Create a migration:

**For PostgreSQL:**
```bash
make create-pg name=create_doctors_table
```

**For MySQL:**
```bash
make create-my name=create_doctors_table
```

Then edit the generated migration file in `migrations/postgres/` or `migrations/mysql/`:

```sql
-- migrations/postgres/xxxx_create_doctors_table.up.sql
CREATE TABLE doctors (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    specialty VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_doctors_email ON doctors(email);
```

Apply the migration:
```bash
make pg-up    # or make my-up for MySQL
```

---

### Step 4: Define SQL Queries

Open `internal/modules/doctor/repository/queries.sql` and write your SQL queries using sqlc syntax:

```sql
-- name: FindDoctorByID :one
SELECT id, name, specialty, email, created_at 
FROM doctors 
WHERE id = $1 AND deleted_at IS NULL;

-- name: InsertDoctor :exec
INSERT INTO doctors (id, name, specialty, email, created_at) 
VALUES ($1, $2, $3, $4, NOW());

-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (id, topic, payload, status, created_at) 
VALUES ($1, $2, $3, 'pending', NOW());
```

📝 **Query directives:**
- `:one` → returns one row
- `:many` → returns multiple rows
- `:exec` → executes without returning rows

---

### Step 5: Generate SQL Code with sqlc

Run sqlc to auto-generate Go code from your SQL queries:

```bash
sqlc generate
```

This creates:
- `internal/modules/doctor/repository/query/querier.go` → Interface
- `internal/modules/doctor/repository/query/queries.sql.go` → Implementation
- `internal/modules/doctor/repository/query/models.go` → Row structs

✅ **Check:** Open the generated files to see what was created. You'll use these in the repository.

---

### Step 6: Implement Repository Layer

Open `internal/modules/doctor/repository/doctor.go` and implement the repository:

```go
package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"  // or mysql for MySQL modules
	v1 "github.com/sabiqazhar/clinic-monolith/contracts/events/v1"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/domain"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/repository/query"
	"go.uber.org/zap"
)

type pgRepo struct {
	db  *pgxpool.Pool
	q   query.Querier
	log *zap.Logger
}

func NewDoctorRepo(db *pgxpool.Pool, log *zap.Logger) domain.DoctorRepository {
	return &pgRepo{
		db:  db,
		q:   query.New(db),
		log: log,
	}
}

func (r *pgRepo) FindByID(ctx context.Context, id string) (*domain.Doctor, error) {
	row, err := r.q.FindDoctorByID(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, domain.ErrDoctorNotFound
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &domain.Doctor{
		ID:        row.ID,
		Name:      row.Name,
		Specialty: row.Specialty,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *pgRepo) SaveWithOutbox(ctx context.Context, d *domain.Doctor) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	qTx := query.New(tx)

	if err := qTx.InsertDoctor(ctx, query.InsertDoctorParams{
		ID:        d.ID,
		Name:      d.Name,
		Specialty: d.Specialty,
		Email:     d.Email,
	}); err != nil {
		return fmt.Errorf("insert doctor failed: %w", err)
	}

	payload, _ := json.Marshal(v1.DoctorRegisteredV1{
		DoctorID:  d.ID,
		Name:      d.Name,
		Specialty: d.Specialty,
		Email:     d.Email,
	})

	if err := qTx.InsertOutboxEvent(ctx, query.InsertOutboxEventParams{
		ID:      uuid.New().String(),
		Topic:   "app.doctor.registered.v1",
		Payload: payload,
	}); err != nil {
		return fmt.Errorf("insert outbox failed: %w", err)
	}

	return tx.Commit(ctx)
}
```

💡 **Pattern:** This uses transactional outbox—saving data + queuing events in one transaction.

---

### Step 7: Implement Service Layer

Open `internal/modules/doctor/service/doctor.go`:

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/domain"
	"go.uber.org/zap"
)

type doctorService struct {
	repo  domain.DoctorRepository
	cache domain.CacheManager
	pub   domain.EventPublisher
	log   *zap.Logger
}

func NewDoctorService(
	repo domain.DoctorRepository,
	cache domain.CacheManager,
	pub domain.EventPublisher,
	log *zap.Logger,
) domain.DoctorService {
	return &doctorService{
		repo:  repo,
		cache: cache,
		pub:   pub,
		log:   log,
	}
}

func (s *doctorService) GetProfile(ctx context.Context, id string) (*domain.Doctor, error) {
	cacheKey := fmt.Sprintf("app:doctor:profile:%s", id)

	// 1. Try cache first
	if data, err := s.cache.Get(ctx, cacheKey); err == nil {
		var d domain.Doctor
		if err := json.Unmarshal(data, &d); err == nil {
			return &d, nil
		}
	}

	// 2. Cache miss → query DB
	doctor, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("doctor lookup failed: %w", err)
	}

	// 3. Async cache warming (fire & forget)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, marshalErr := json.Marshal(doctor)
		if marshalErr != nil {
			s.log.Error("fail to marshal doctor for cache", zap.Error(marshalErr))
			return
		}

		_ = s.cache.Set(bgCtx, cacheKey, data, 15*time.Minute)
	}()

	return doctor, nil
}

func (s *doctorService) Register(ctx context.Context, name, specialty, email string) (*domain.Doctor, error) {
	d := &domain.Doctor{
		ID:        uuid.New().String(),
		Name:      name,
		Specialty: specialty,
		Email:     email,
		CreatedAt: time.Now(),
	}

	if err := s.repo.SaveWithOutbox(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to register doctor: %w", err)
	}

	return d, nil
}
```

💡 **Note:** The service doesn't know about SQL—it only talks to interfaces!

---

### Step 8: Implement Handler Layer

Open `internal/modules/doctor/handler/doctor.go`:

```go
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/domain"
	"go.uber.org/zap"
)

type DoctorHandler struct {
	svc domain.DoctorService
	log *zap.Logger
}

func NewDoctorHandler(svc domain.DoctorService, log *zap.Logger) *DoctorHandler {
	return &DoctorHandler{svc: svc, log: log}
}

// RegisterRoutes tells Gin which routes this handler handles
func (h *DoctorHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/:id", h.GetProfile)
	g.POST("/", h.Register)
}

// GET /api/v1/doctors/:id
func (h *DoctorHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing doctor id"})
		return
	}

	ctx := c.Request.Context()
	doctor, err := h.svc.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
			return
		}
		h.log.Error("failed to get doctor profile",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": doctor})
}

// POST /api/v1/doctors
func (h *DoctorHandler) Register(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Specialty string `json:"specialty" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()
	doctor, err := h.svc.Register(ctx, req.Name, req.Specialty, req.Email)
	if err != nil {
		h.log.Error("failed to register doctor",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register doctor"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": doctor})
}
```

---

### Step 9: Configure Dependency Injection (provider.go)

Open `internal/modules/doctor/provider.go`:

```go
package doctor

import (
	"github.com/goforj/wire"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/handler"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/repository"
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/service"
)

var DoctorSet = wire.NewSet(
	repository.NewDoctorRepo,
	service.NewDoctorService,
	handler.NewDoctorHandler,
)
```

💡 **What this does:** Tells Wire how to build your module's dependencies.

---

### Step 10: Register Module in wire.go

Open `cmd/api/wire.go` and:

1. **Add imports** at the top:
```go
import (
	// ... existing imports ...
	"github.com/sabiqazhar/clinic-monolith/internal/modules/doctor"
	doctordomain "github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/domain"
	doctorhandler "github.com/sabiqazhar/clinic-monolith/internal/modules/doctor/handler"
)
```

2. **Add to App struct**:
```go
type App struct {
	PatientHandler     *patienthandler.PatientHandler
	BillingHandler     *billinghandler.BillingHandler
	AppointmentHandler *appointmenthandler.AppointmentHandler
	PatientSubscriber  *billingsubscriber.PatientSubscriber
	DoctorHandler      *doctorhandler.DoctorHandler  // ← Add this
}
```

3. **Add interface bindings** (inside `InitializeApp`, before `wire.Build`):
```go
// Doctor domain interfaces
wire.Bind(new(doctordomain.CacheManager), new(*cacheAdapter)),
wire.Bind(new(doctordomain.EventPublisher), new(*publisherAdapter)),
```

4. **Add provider set** (inside `wire.Build`):
```go
doctor.DoctorSet,  // ← Add this line
```

---

### Step 11: Mount Routes in main.go

Open `cmd/api/main.go` and find where routes are mounted (around line 145):

```go
// Mount routes per modul (Handler Self-Registration Pattern)
app.PatientHandler.RegisterRoutes(v1.Group("/patients"))
app.AppointmentHandler.RegisterRoutes(v1.Group("/appointments"))
app.BillingHandler.RegisterRoutes(v1.Group("/billing"))
app.DoctorHandler.RegisterRoutes(v1.Group("/doctors"))  // ← Add this line
```

---

### Step 12: Generate Wire Code

Now generate the dependency injection code:

```bash
wire ./cmd/api
```

This creates/updates `cmd/api/wire_gen.go`.

✅ **Success check:** If you see errors, read them carefully—usually it's a missing import or mismatched interface.

---

### Step 13: Build and Run

Finally, build and test your module:

```bash
# Build
go build ./cmd/api

# Run
go run ./cmd/api
```

Test your endpoints:

```bash
# Register a new doctor
curl -X POST http://localhost:8080/api/v1/doctors \
  -H "Content-Type: application/json" \
  -d '{"name":"Dr. Smith","specialty":"Cardiology","email":"dr.smith@clinic.com"}'

# Get doctor profile
curl http://localhost:8080/api/v1/doctors/<doctor-id>
```

---

### Quick Checklist

Before you're done, verify:

- [ ] Module scaffolded with `make create-module`
- [ ] Domain interfaces defined
- [ ] Database migration created and applied
- [ ] SQL queries written in `queries.sql`
- [ ] `sqlc generate` ran successfully
- [ ] Repository implemented
- [ ] Service implemented
- [ ] Handler implemented with `RegisterRoutes`
- [ ] `provider.go` configured
- [ ] Module imported in `cmd/api/wire.go`
- [ ] Interface bindings added in `wire.go`
- [ ] Provider set added in `wire.Build`
- [ ] Handler added to `App` struct
- [ ] Routes mounted in `cmd/api/main.go`
- [ ] `wire ./cmd/api` generated successfully
- [ ] App builds and runs without errors
- [ ] Endpoints tested with curl/Postman

---

### Common Pitfalls & Tips

🚨 **"no rows in result set" error:** Make sure your migration was applied (`make pg-version` to check).

🚨 **Wire compilation errors:** Usually means a missing binding or wrong interface. Check `wire.go` bindings.

🚨 **SQLC errors:** Verify your `queries.sql` syntax matches the engine (PostgreSQL uses `$1`, MySQL uses `?`).

💡 **Copy-Paste Strategy:** The easiest way to start is to copy an existing module (like `patient`) and rename everything. Then modify as needed.

💡 **Ask for Help:** Stuck? Look at existing modules (`patient`, `billing`, `appointment`)—they're your best reference!

---


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
