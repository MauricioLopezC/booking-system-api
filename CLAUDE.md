# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run API server
go run ./cmd/api

# Build binaries
go build -o reservation-api ./cmd/api
go build -o seed ./cmd/seed

# Tests
go test ./...                           # all tests
go test -v ./...                        # verbose
go test -cover ./...                    # with coverage
go test -run TestGetSlots ./internal/slots   # single test by name

# Integration tests (skipped unless TEST_DATABASE_DSN is set).
# They DROP the public schema of that database: use a throwaway one.
docker run -d --rm --name reservation-test-db -e POSTGRES_PASSWORD=test -e POSTGRES_DB=reservation_test -p 127.0.0.1:55432:5432 postgres:17-alpine
TEST_DATABASE_DSN="host=127.0.0.1 port=55432 user=postgres password=test dbname=reservation_test sslmode=disable" go test -p 1 ./...

# Seed database
go run ./cmd/seed           # seed with default data
go run ./cmd/seed -reset    # truncate all tables then seed

# Generate OpenAPI docs (swagger)
swag init --generalInfo cmd/api/main.go --dir . --output docs
```

## Environment

Create a `.env` file at the project root with these required variables:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=<postgres_user>
DB_PASSWORD=<postgres_password>
DB_NAME=reservation-system
SECRET=<jwt_secret_key>
```

All 6 variables are mandatory — startup fails if any are missing. Optional:
- `SLOT_STEP_MINUTES` (default 15): grid between offered slot start times.
- `BUSINESS_TIMEZONE` (default `UTC`, IANA name like `America/Argentina/Buenos_Aires`): schedules and exceptions are wall-clock times in this zone; slot days are local dates; instants are stored and returned in UTC. The process timezone (`time.Local`) is forced to UTC at startup because postgres `time` columns are read in `time.Local`; this is independent of `BUSINESS_TIMEZONE`.

## Architecture

REST API for a booking/reservation system. Stack: **Fiber v3** (HTTP), **GORM** (ORM), **PostgreSQL**, JWT auth, Swagger at `/swagger/*`.

**Layered pattern:** Handler → Service → Database (GORM). Services and handlers are instantiated in `cmd/api/main.go` and wired via initialization functions.

**Key packages under `internal/`:**
- `auth/` — JWT login/register, bcrypt password hashing
- `user/` — User CRUD, roles: `admin | employee | customer`
- `bookings/` — Reservation management; `EndDate` is stored at creation time; the `bookings_no_overlap` exclusion constraint (postgres, `btree_gist`) prevents overlapping active bookings per employee; max 14 days ahead
- `schedule/` — Weekly recurring employee schedules (weekday + start/end time)
- `schedule_exceptions/` — Overrides to base schedule (all-day or partial time blocks, with `IsAvailable` flag)
- `slots/` — **Core business logic**: per day, free time = (weekly schedule + `IsAvailable` exceptions) − (blocking exceptions + active bookings), computed with pure interval functions in `intervals.go`. Each free interval offers its own start and then points of a `SLOT_STEP_MINUTES` grid where the service fits. Most heavily tested package.
- `services/` — Haircut/beauty services (name, duration_minutes, price); many-to-many with User
- `model/` — All GORM model definitions and enums in one file
- `common/` — Shared error types (`ErrRecordNotFound`, `ErrValidation`, `ErrConflict`, `ErrUnauthorized`) and response structs
- `utils/` — `TimeOfDay` (date-agnostic time value), validation helpers, time parsing

**Database migrations** run automatically on every startup via `database.Migrate(db)` (AutoMigrate plus idempotent raw SQL for what AutoMigrate can't express, e.g. the exclusion constraint). The seed uses the same function.

**Access control** (defined in `cmd/api/router.go`):
- Public: `GET /api/services`, `GET /api/available-slots`
- Authenticated: `/api/me/*` routes
- Admin: `POST/PATCH/DELETE /api/services`, `/api/bookings`, `/api/schedules`
- Employee: schedule management and service assignment

**Rate limiting:** 60 req/min general, 5 req/min on auth endpoints.

## Testing

Tests use `stretchr/testify` with table-driven test cases. Coverage focuses on schedule overlap detection, slot generation edge cases (exceptions, existing bookings, past times), and `TimeOfDay` parsing. No mocking — business logic is tested as pure functions; DB-level guarantees (overlap constraint, concurrency, migrations) are covered by integration tests using `testutil.OpenTestDB`.
