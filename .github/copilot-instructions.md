# Copilot Instructions - Kubernetes Multi-Tier App

## Project Overview

Go-based microservice designed for Kubernetes deployment with PostgreSQL backend. Implements a **Quick Poll System** API with voting capabilities. Monorepo structure with separate `server/`, `client/`, `k8s/`, and `charts/` directories.

## Architecture Principles

### Server Structure (`server/`)

- **Module**: `github.com/moabdelazem/k8s-app` (note: repo name differs from module path)
- **Entry point**: `cmd/main.go` - initializes config → logger → database → router → HTTP server
- **Clean Architecture layers**:
  - `internal/models/` - Domain entities (Poll, PollOption, Vote)
  - `internal/repository/` - Database access layer with SQL queries
  - `internal/service/` - Business logic, validation, orchestration
  - `internal/api/handlers/` - HTTP request/response handling
  - `internal/api/router.go` - Route definitions and middleware
  - `internal/config/` - Environment-based configuration
  - `internal/database/` - PostgreSQL connection pool with retry logic
  - `pkg/` - Reusable utilities (logger, response, env)

### Layer Communication Pattern

**Handler → Service → Repository → Database**

- **Handlers**: Parse HTTP requests, call services, format responses
- **Services**: Business logic, validation, transaction coordination
- **Repositories**: SQL queries, database operations
- **Never skip layers**: Handlers must not call repositories directly

Example dependency injection in `router.go`:

```go
pollRepo := repository.NewPollRepository(db)
pollService := service.NewPollService(pollRepo)
pollHandler := handlers.NewPollHandler(pollService)
```

### Configuration Pattern

- Uses `godotenv` to load `.env` files automatically in `config.NewConfig()`
- Default port: **6767** (not 8080) as defined in `.env.example`
- ENV variable controls logger behavior: `development` (console, colored) vs `production` (JSON)
- Config includes DB connection pool settings AND retry configuration
- Config validation happens at initialization, not lazily

### Database Connection Pattern

- **Driver**: `lib/pq` for PostgreSQL
- **Global instance**: `database.DB` initialized in `main()` after logger
- **Retry logic with exponential backoff**: Configurable via `DB_MAX_RETRIES` and `DB_RETRY_DELAY`
- Connection pool configured via: `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`
- Always `defer database.Close()` in `main()`
- Use `database.Ping()` for health checks, `database.Stats()` for pool metrics
- Example: 5 retries with 2s initial delay, exponentially increasing per attempt

### Transaction Pattern

All multi-step database operations use transactions:

```go
tx, err := r.db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback() // Auto-rollback if commit not reached

// Multiple operations...
err = tx.QueryRowContext(ctx, query, args...).Scan(...)
if err != nil {
    return err // Rollback happens automatically
}

return tx.Commit()
```

### Logging with Zap

- **Global logger**: `logger.Log` initialized in `main()` before any other operations
- **Environment-aware**: Development uses colored console output, production uses JSON
- **Structured fields required**: Always use `zap.String()`, `zap.Error()`, etc., not string interpolation
- **Always defer**: `defer logger.Sync()` immediately after initialization
- Example: `logger.Info("Vote cast", zap.String("poll_id", id.String()), zap.String("voter", identifier))`
- Log at service layer for business events, handler layer for request/response events

### HTTP Response Pattern

- **Never write raw bytes**: Use `pkg/response` helpers exclusively
- Standard envelope: `{"success": bool, "message": string, "data": any, "error": string}`
- Common helpers: `response.Success()`, `response.Created()`, `response.BadRequest()`, `response.NotFound()`
- For custom status: `response.JSON(w, statusCode, data)`

### Router & Middleware (Chi v5)

- Middleware stack in `router.go`: RequestID → Recoverer → LoggingMiddleware
- Routes organized with `r.Route()` for grouping (e.g., `/api/v1/polls`)
- Handler registration requires database instance: `SetupRoutes(db *sql.DB)`
- URL parameters extracted with: `chi.URLParam(r, "id")`
- Health endpoints for K8s: `/health` (detailed with DB stats), `/live` (liveness), `/ready` (readiness with DB ping)

## Poll System Implementation

### Database Schema

- **polls**: Question, description, expiration, vote count
- **poll_options**: Options with vote counts, ordered by position
- **votes**: Individual votes with unique constraint per voter per poll
- **Triggers**: Automatically update poll total_votes on vote insert/delete
- **Voter identification**: Uses IP address (X-Forwarded-For → X-Real-IP → RemoteAddr)

### API Endpoints

```
POST   /api/v1/polls              # Create poll (2-10 options required)
GET    /api/v1/polls               # List polls (pagination: ?limit=20&offset=0&active=true)
GET    /api/v1/polls/:id           # Get poll with results and percentages
POST   /api/v1/polls/:id/vote      # Vote on poll (one vote per voter)
DELETE /api/v1/polls/:id           # Soft delete (sets is_active=false)
```

### Validation Rules

- Question: 5-500 characters
- Options: 2-10 options, each 1-200 characters
- Expiration: Must be future date if provided
- Voting: Poll must be active and not expired
- Duplicate prevention: Unique constraint on (poll_id, voter_identifier)

### Concurrency Handling

- **Atomic vote counting**: Database triggers maintain vote counts
- **Transactions**: Create poll + options in single transaction
- **Race condition prevention**: Unique constraint prevents duplicate votes
- **Lock-free reads**: Vote counts updated by triggers, no application-level locking

## Development Workflow

### Essential Commands (from `server/`)

```bash
make run      # Development server (go run, no rebuild on change)
make build    # Compiles to bin/app
make test     # Run all tests

# Database
docker compose -f compose.dev.yaml up -d    # Start PostgreSQL 16
docker compose -f compose.dev.yaml down     # Stop and remove
docker compose -f compose.dev.yaml logs -f  # Follow logs
```

### Database Setup

- **PostgreSQL 16 Alpine** via `compose.dev.yaml`
- Credentials in `.env`: `devuser:devpassword@localhost:5432/k8s_app_dev`
- Init scripts in `init-scripts/` run automatically on first container creation
- Includes: uuid-ossp extension, polls tables, indexes, triggers
- Named volume `postgres_dev_data` persists between container restarts
- Container name: `k8s_app_postgres_dev`, network: `k8s_app_network`

### Initialization Order (Critical)

1. Load config (automatically loads `.env`)
2. Initialize logger (before any logging calls)
3. Initialize database connection (with retry logic)
4. Setup router with database instance
5. Start HTTP server

### Adding Dependencies

```bash
cd server
go get package-name    # Add to go.mod
go mod tidy           # Clean up
```

## Code Conventions

### Adding New Feature (Full Stack)

1. **Model** (`internal/models/`): Define structs with JSON tags
2. **Repository** (`internal/repository/`): Create `*Repository` struct with `New*Repository(db)` constructor
3. **Service** (`internal/service/`): Create `*Service` struct with `New*Service(repo)` constructor
4. **Handler** (`internal/api/handlers/`): Create `*Handler` struct with `New*Handler(service)` constructor
5. **Router** (`internal/api/router.go`): Wire dependencies and register routes

### Handler Structure

```go
func (h *Handler) MethodName(w http.ResponseWriter, r *http.Request) {
    logger.Info("Handler started", zap.String("handler", "MethodName"))

    // 1. Parse/validate input
    var req RequestModel
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "Invalid request body")
        return
    }

    // 2. Call service
    result, err := h.service.DoSomething(r.Context(), &req)
    if err != nil {
        logger.Error("Operation failed", zap.Error(err))
        response.InternalServerError(w, err.Error())
        return
    }

    // 3. Return response
    response.Success(w, "Operation successful", result)
}
```

### Service Pattern

```go
func (s *Service) Method(ctx context.Context, input Type) (*Output, error) {
    // 1. Validate business rules
    if err := validateInput(input); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // 2. Call repository
    result, err := s.repo.DatabaseOperation(ctx, input)
    if err != nil {
        logger.Error("Database operation failed", zap.Error(err))
        return nil, fmt.Errorf("operation failed: %w", err)
    }

    // 3. Log success
    logger.Info("Operation completed", zap.String("id", result.ID.String()))
    return result, nil
}
```

### Repository Pattern

```go
func (r *Repository) Method(ctx context.Context, param Type) (*Result, error) {
    query := `SELECT ... FROM table WHERE id = $1`

    var result Result
    err := r.db.QueryRowContext(ctx, query, param).Scan(&result.Field1, &result.Field2)
    if err == sql.ErrNoRows {
        return nil, nil // Or appropriate not-found handling
    }
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }

    return &result, nil
}
```

### UUID Handling

- Use `github.com/google/uuid` package
- Parse from string: `uuid.Parse(str)` - always check error
- Generate: `uuid.New()` or database `uuid_generate_v4()`
- In JSON: UUIDs serialize as strings automatically

### Context Usage

- Always accept `context.Context` as first parameter in service/repository methods
- Pass `r.Context()` from handlers
- Use for cancellation, timeouts, and request-scoped values

## Health Check Pattern

- `/health`: Returns detailed system info including database connection pool stats
- `/live`: Simple liveness probe (returns alive status)
- `/ready`: Readiness probe that pings database - returns 503 if DB unhealthy
- Health endpoints use `database.Ping()` and `database.Stats()` to check DB status
- Connection pool stats include: OpenConnections, InUse, Idle, WaitCount, WaitDuration, MaxIdleClosed, MaxLifetimeClosed

## Critical Details

- **Handler naming**: Files like `poll_handler.go`, functions like `CreatePoll()` (not `CreatePollHandler`)
- **Import path**: Always `github.com/moabdelazem/k8s-app/...` regardless of repo name
- **Makefile from server/**: All make commands must run from `server/` directory, not repo root
- **Database retry**: App will retry connection 5 times (default) with exponential backoff before failing
- **Router requires DB**: `SetupRoutes(db)` needs database instance for dependency injection
- **Soft deletes**: Use `is_active` flag, don't hard delete from database
- **Pagination defaults**: limit=20, max=100 to prevent resource exhaustion

## Common Patterns to Follow

1. **New API endpoint**: Model → Repository → Service → Handler → Router (full stack)
2. **New config field**: Add to `Config` struct → Update `.env.example` and `.env` → Add validation
3. **Logging**: Use structured fields at service/handler layers: `zap.String()`, `zap.Int()`, `zap.Error()`
4. **Error handling**: Wrap errors with context using `fmt.Errorf("context: %w", err)`
5. **Database queries**: Always use `QueryRowContext` or `QueryContext` with context parameter
6. **Transactions**: Use `defer tx.Rollback()` immediately after `BeginTx()`

## Kubernetes Readiness

- Health endpoints designed for K8s probes (`/live` for liveness, `/ready` for readiness)
- Container name pattern: `k8s_app_postgres_dev` (prefix with `k8s_app_`)
- Docker network: `k8s_app_network` for service discovery
- Database retry logic ensures graceful startup when DB isn't ready immediately
- ReadinessProbe accurately reflects DB connection status via ping
