# Event Booking System

An event booking system with limited capacity. Bookings never exceed capacity under concurrency. The system supports booking, cancellation (returns the spot to capacity), and audit logging of every booking-changing operation.

**Stack:** Go (gorilla/mux, database/sql, lib/pq) | PostgreSQL 16 | React 18 (Vite) | Docker Compose

---

## Table of Contents

1. [How It Works](#how-it-works)
2. [Running the Application](#running-the-application)
3. [API Reference](#api-reference)
4. [Testing](#testing)
5. [Database Schema and Migrations](#database-schema-and-migrations)
6. [Concurrency and Transaction Design](#concurrency-and-transaction-design)
7. [Project Structure](#project-structure)
8. [Assumptions](#assumptions)
9. [Design Decisions](#design-decisions)

---

## How It Works

The application has pre-seeded users and events. There is no authentication and no event management UI. Users are selected from a dropdown in the frontend, and all operations happen in the context of that selected user.

**Booking flow:**
A user picks an event and clicks "Book Now." The backend atomically checks capacity and creates the booking in a single transaction. If the event is full, the user sees "Sold Out." If they already have an active booking for that event, they see "Already Booked." On success, a confirmation popup appears and the button disables.

**Cancellation flow:**
From the "My Bookings" page, a user cancels an active booking. The backend atomically marks the booking as cancelled and decrements the event's booked count, returning the spot to capacity. The cancel is scoped to the requesting user -- a user cannot cancel someone else's booking.

**Audit log:**
Every booking-changing operation (book and cancel) is recorded in an immutable audit log table. Successful operations are logged inside the same transaction as the booking change. Failed operations (sold out, already booked, already cancelled) are logged in a separate database call so the record persists even though the main transaction rolls back. The audit log page supports filtering by event, user, operation type, and outcome.

---

## Running the Application

There are two ways to run the system: **Docker Compose** (recommended, single command) or **Local** (separate processes, useful for development).

### Option 1: Docker Compose

This starts three containers: PostgreSQL, the Go backend, and Nginx serving the React frontend.

```bash
# Start everything
make docker-up

# The app is available at:
#   Frontend: http://localhost
#   Backend:  http://localhost:8080
```

**Logs:**

```bash
make compose-logs
```

**Cleanup (removes containers, volumes, and images):**

```bash
make clean-docker
```

### Option 2: Local Setup

This runs a standalone PostgreSQL container, the Go backend via `go run`, and the Vite dev server for the frontend. Useful when you want hot-reload on the frontend and faster iteration on the backend.

**Prerequisites:** Docker (for PostgreSQL), Go 1.24+, Node.js 20+.

```bash
# One-time setup: starts Postgres, creates user/database,
# runs migrations, installs frontend npm dependencies.
make setup
```

Then open **two terminals:**

```bash
# Terminal 1 -- backend (port 8080)
make backend

# Terminal 2 -- frontend with hot reload (port 3000, proxies /api to backend)
make frontend
```

The app is available at http://localhost:3000.

**Logs:**

Backend logs are written to `./event-booking.log` by default. To tail them:

```bash
make logs
```

To change the log destination or level:

```bash
make backend LOG_OUTPUT=stdout LOG_LEVEL=debug
```

**Cleanup (removes container, Postgres data, node_modules, log file):**

```bash
make clean-local
```

**Reset database without full cleanup:**

```bash
make reset
```

The Makefile has additional targets for inspecting the database, rebuilding individual Docker services, and more. Run `make help` to see all available commands.

---

## API Reference

All responses follow a standard envelope:

```json
{ "success": true, "data": { ... }, "message": "..." }
{ "success": false, "error": "...", "message": "..." }
```

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Liveness probe |
| GET | `/api/users` | List all pre-seeded users |
| GET | `/api/events` | List all events with capacity info |
| GET | `/api/events/{id}` | Single event detail |
| POST | `/api/events/{id}/book` | Book a spot (body: `{"user_id": "uuid"}`) |
| POST | `/api/bookings/{id}/cancel` | Cancel a booking (body: `{"user_id": "uuid"}`) |
| GET | `/api/users/{id}/bookings` | List user's bookings (optional `?status=active`) |
| GET | `/api/audit` | Audit log (optional filters: `event_id`, `user_id`, `booking_id`, `operation`, `outcome`) |

---

## Testing

### Manual testing with curl

```bash
# List events
curl http://localhost:8080/api/events

# Alice books the Go Workshop (capacity: 3)
curl -X POST http://localhost:8080/api/events/e1111111-1111-1111-1111-111111111111/book \
  -H "Content-Type: application/json" \
  -d '{"user_id": "a1111111-1111-1111-1111-111111111111"}'

# Cancel that booking (replace <booking-id> with the id from the response above)
curl -X POST http://localhost:8080/api/bookings/<booking-id>/cancel \
  -H "Content-Type: application/json" \
  -d '{"user_id": "a1111111-1111-1111-1111-111111111111"}'

# View audit logs
curl http://localhost:8080/api/audit

# View audit logs filtered by event
curl "http://localhost:8080/api/audit?event_id=e1111111-1111-1111-1111-111111111111"
```

### Load test tool

A standalone Go CLI tool lives at `backend/cmd/loadtest/`. It proves that the concurrency model is correct by firing many goroutines at the same event simultaneously and verifying the results. All goroutines use the **start-gun pattern** -- they block on a shared channel and fire at the exact same moment when the channel is closed, creating maximum contention.

The tool communicates with the backend over HTTP (same API the frontend uses), so it tests the full stack end-to-end: HTTP handler, service layer, repository transactions, and PostgreSQL row locking.

**Prerequisites:** The backend must be running (either via `make backend` or `make docker-up`).

**Three test scenarios:**

**1. Booking test** -- Proves bookings never exceed capacity.

Fires N goroutines that all try to book the same event at once. With a capacity-3 event and 50 goroutines, exactly 3 succeed and 47 get "sold out." The tool fetches the event afterward and verifies `booked_count == capacity`.

```bash
cd backend

# Default: 50 goroutines against Go Workshop (capacity 3)
go run ./cmd/loadtest --test book

# Custom goroutines and event
go run ./cmd/loadtest --test book --goroutines 100 --event-id e8888888-8888-8888-8888-888888888888
```

Expected output:

```
Event:      Go Workshop 2025 (capacity: 3, booked: 0, remaining: 3)
Goroutines: 50
Test:       book

Completed in 143ms

Results:
  Successful bookings:  3
  Sold out:             47
  Already booked:       0
  Errors:               0

Verification:
  Expected booked:  3
  Actual booked:    3
  Capacity:         3

PASS -- no overbooking, no errors
```

**2. Cancel test** -- Proves cancellation returns spots atomically and count never goes negative.

First books the event to full capacity sequentially, then fires all cancellations simultaneously. Verifies `booked_count` returns to 0 and never goes below zero.

```bash
go run ./cmd/loadtest --test cancel --event-id e3333333-3333-3333-3333-333333333333
```

Expected output:

```
Event:      System Design Talk (capacity: 2, booked: 0)
Test:       cancel

Booking 2 spots sequentially...
Created 2 bookings

Firing 2 concurrent cancellations...

Results:
  Successful cancels:   2
  Not found/duplicate:  0
  Errors:               0

Verification:
  Booked count after:   0

PASS -- no negative count, no errors
```

**3. Mixed test** -- Proves the system stays consistent under simultaneous cancels and new bookings.

Books the event to capacity, then fires N goroutines: half cancel existing bookings while the other half try to book freed spots. This is the hardest scenario -- cancellations free capacity that booking goroutines race to claim. The tool verifies `booked_count` stays within `[0, capacity]`.

```bash
go run ./cmd/loadtest --test mixed --event-id e8888888-8888-8888-8888-888888888888
```

Unlike the book and cancel tests, the mixed test output is **non-deterministic**. The cancel and booking goroutines race against each other, and the number of successful rebooks depends on OS thread scheduling and PostgreSQL lock acquisition order. One run may show 2 rebooks, another may show 4. The only guarantee is the invariant: `0 <= booked_count <= capacity`. That is what PASS/FAIL checks.

Sample output (numbers will vary between runs):

```
Event:      PostgreSQL Internals (capacity: 5, booked: 0)
Test:       mixed (concurrent cancel + rebook)

Booking 5 spots sequentially...
Created 5 bookings, event is now full

Firing 10 goroutines: 5 cancels + 5 booking attempts...

Completed in 19ms

Results:
  Successful cancels:   5
  Successful rebooks:   2       <-- varies between runs

Verification:
  Booked count after:   2       <-- varies, but always within [0, 5]
  Capacity:             5

PASS -- booked_count is within [0, capacity]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--base-url` | `http://localhost:8080` | Backend API URL |
| `--event-id` | `e1111111-...` (Go Workshop) | Event UUID to target |
| `--goroutines` | `50` | Number of concurrent goroutines (book test only) |
| `--test` | `book` | Test type: `book`, `cancel`, or `mixed` |

Note: after running load tests the targeted events will have bookings. Run `make reset` (local) or `docker-compose down -v && docker-compose up -d` (Docker) to reset the database to a clean state.

### Inspecting the database

```bash
# Open a psql shell
make db-shell

# Show all tables
make db-tables

# Show all indexes
make db-indexes
```

---

## Database Schema and Migrations

Migrations live in `backend/migrations/` and run automatically on server startup (in Docker) or via `make db-migrate` (local setup). They are numbered and executed in lexical order. Each migration is idempotent -- tables use `IF NOT EXISTS` and seed data uses `ON CONFLICT DO NOTHING` -- so re-running them is safe.

| File | Purpose |
|------|---------|
| `001_create_users.sql` | Users table with UUID primary key and unique email |
| `002_create_events.sql` | Events table with `capacity`, `booked_count`, and CHECK constraints |
| `003_create_bookings.sql` | Bookings table with partial unique index preventing duplicate active bookings |
| `004_create_audit_logs.sql` | Immutable audit log table with nullable `booking_id` and `failure_reason` |
| `005_seed_data.sql` | Pre-seeded users and events with fixed UUIDs for reproducible testing |

### Pre-seeded users

| Name | UUID | Email |
|------|------|-------|
| Alice | `a1111111-1111-1111-1111-111111111111` | alice@example.com |
| Bob | `b2222222-2222-2222-2222-222222222222` | bob@example.com |
| Charlie | `c3333333-3333-3333-3333-333333333333` | charlie@example.com |
| Dave | `d4444444-4444-4444-4444-444444444444` | dave@example.com |
| Eve | `a5555555-5555-5555-5555-555555555555` | eve@example.com |
| Frank | `a6666666-6666-6666-6666-666666666666` | frank@example.com |
| Grace | `a7777777-7777-7777-7777-777777777777` | grace@example.com |
| Heidi | `a8888888-8888-8888-8888-888888888888` | heidi@example.com |
| Ivan | `a9999999-9999-9999-9999-999999999999` | ivan@example.com |

### Pre-seeded events

| Title | UUID | Capacity | Location |
|-------|------|----------|----------|
| Go Workshop 2025 | `e1111111-1111-1111-1111-111111111111` | 3 | Noida |
| React Meetup | `e2222222-2222-2222-2222-222222222222` | 20 | Delhi |
| System Design Talk | `e3333333-3333-3333-3333-333333333333` | 2 | Bangalore |
| AI/ML Hackathon | `e4444444-4444-4444-4444-444444444444` | 50 | Noida |
| DevOps Bootcamp | `e5555555-5555-5555-5555-555555555555` | 10 | Gurugram |
| Rust for Beginners | `e6666666-6666-6666-6666-666666666666` | 15 | Hyderabad |
| Cloud Native Day | `e7777777-7777-7777-7777-777777777777` | 30 | Pune |
| PostgreSQL Internals | `e8888888-8888-8888-8888-888888888888` | 5 | Bangalore |
| Frontend Performance | `e9999999-9999-9999-9999-999999999999` | 25 | Delhi |

Events with low capacity (2, 3, and 5) are intentional -- they make it easy to test the sold-out flow manually.

---

## Concurrency and Transaction Design

This is the most critical part of the system. The core requirement is: bookings must never exceed capacity, even when many users book the same event at the same time.

### Booking: Atomic Conditional Update

Three approaches were evaluated for preventing overbooking under concurrency.

**Approach 1: Pessimistic Locking (SELECT FOR UPDATE)**

```sql
BEGIN TX
  SELECT * FROM events WHERE id = $1 FOR UPDATE  -- row lock acquired here
  -- application checks: IF booked_count >= capacity THEN ROLLBACK
  INSERT INTO bookings (event_id, user_id, status) VALUES ($1, $2, 'active')
  UPDATE events SET booked_count = booked_count + 1
COMMIT
```

The `FOR UPDATE` clause acquires an exclusive row lock when the SELECT executes. Any other transaction trying to SELECT FOR UPDATE on the same row blocks until this transaction commits or rolls back. This is correct -- no two transactions can read and modify the row at the same time. The downside is throughput: every booking request for the same event queues behind the lock, even if the event has plenty of capacity. Under high concurrency (e.g., a popular event with 1000 users hitting "Book Now" simultaneously), the serialization becomes a bottleneck.

**Approach 2: Optimistic Locking (Version/CAS)**

```sql
-- Step 1: read current state (no lock held)
SELECT booked_count, version FROM events WHERE id = $1

-- Step 2: application checks capacity
-- Step 3: attempt update with version guard
UPDATE events
SET booked_count = booked_count + 1, version = version + 1
WHERE id = $1 AND version = $old_version

-- if rows_affected == 0, another transaction modified the row first -- retry
```

No lock is held during the read, so concurrent requests don't block each other. Instead, each transaction reads the current version, does its work, and attempts the update with a version check. If the version changed between the read and the write (because another transaction committed first), the update affects zero rows and the application retries. This gives better throughput than pessimistic locking when contention is low, but under heavy contention many transactions retry repeatedly, wasting CPU and database round-trips. The retry logic with exponential backoff adds complexity.

**Approach 3: Atomic Conditional Update (chosen)**

```sql
UPDATE events
SET booked_count = booked_count + 1, updated_at = NOW()
WHERE id = $1 AND booked_count < capacity
-- if rows_affected == 0 -> event is full (sold out)
-- if rows_affected == 1 -> spot reserved, proceed to insert booking
```

This collapses the read, check, and write into a single SQL statement. PostgreSQL acquires a row-level lock internally during the UPDATE, evaluates the WHERE condition (`booked_count < capacity`), and either applies the change or affects zero rows. The lock is held only for the duration of the UPDATE itself, not across multiple statements. Two concurrent requests targeting the same row serialize at the row lock automatically, but the lock window is minimal compared to Approach 1.

The Go code does not perform any capacity check. It executes the UPDATE, reads `rows_affected`, and interprets the result:

```go
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    return nil, ErrSoldOut  // Postgres told us the event is full
}
// rowsAffected == 1: spot reserved, proceed to insert booking
```

There is no separate SELECT, no version field, no retry loop. The entire concurrency problem is solved by one SQL statement.

**Why Approach 3 was chosen:** It is the simplest correct solution for this use case. Approach 1 is correct but slower under contention. Approach 2 is faster under low contention but adds retry complexity. Approach 3 is both fast and simple -- the only trade-off is that complex pre-update validation (e.g., checking multiple tables before deciding) would require restructuring. For a single capacity check, it is the right tool.

### Booking Transaction

The full booking transaction wraps three operations atomically:
1. Increment `booked_count` (the atomic conditional update above)
2. Insert the booking row
3. Insert the success audit log

If any step fails, the entire transaction rolls back. The database also enforces a CHECK constraint (`booked_count <= capacity`) as an additional safety net at the database level.

A partial unique index on `bookings (event_id, user_id) WHERE status = 'active'` prevents the same user from booking the same event twice at the database level. The application detects this by checking for PostgreSQL error code `23505` (unique violation).

### Cancellation: Atomic Return of Capacity

Cancellation follows the same transactional approach:
1. Mark the booking as `cancelled` (WHERE clause includes `user_id` and `status = 'active'` to prevent cross-user cancellation and double-cancel)
2. Decrement `booked_count` on the event
3. Insert the success audit log

A CHECK constraint (`booked_count >= 0`) prevents the count from going negative, guarding against double-cancel bugs.

### Failure Audit Logging

When a booking or cancellation fails (sold out, already booked, not found), the main transaction is rolled back. The failure audit log is written in a **separate database call** using a fresh connection, so it persists regardless of the rollback. This ensures every attempt -- successful or not -- is recorded.

---

## Project Structure

```
backend/
  cmd/server/main.go           Entry point: config, DB, migrations, routing, server
  internal/
    config/config.go            Environment variable loading
    db/db.go                    PostgreSQL connection pool
    db/migrations.go            Migration runner (reads .sql files in order)
    model/models.go             Domain types (User, Event, Booking, AuditLog) and errors
    repository/interfaces.go    Repository interfaces for dependency injection
    repository/user_repo.go     User queries
    repository/event_repo.go    Event queries
    repository/booking_repo.go  Booking/cancel transactions, failure audit logging
    repository/audit_repo.go    Audit log queries with optional filters
    service/event_service.go    Business logic wrapper over event/user repos
    service/booking_service.go  Business logic wrapper over booking/audit repos
    handler/event_handler.go    HTTP handlers for events
    handler/user_handler.go     HTTP handler for users
    handler/booking_handler.go  HTTP handlers for booking, cancel, user bookings
    handler/audit_handler.go    HTTP handler for audit log with filter parsing
    middleware/cors.go          CORS headers for local development
    middleware/logging.go       Structured JSON request logging
    response/response.go        Standardized JSON response helpers
    logger/logger.go            Structured JSON logger (stdout or file)
  migrations/                   SQL migration files (001-005)
  Dockerfile                    Multi-stage: Go build then Alpine runtime

frontend/
  src/
    api/client.js               Fetch wrapper and API functions
    context/UserContext.jsx      React Context for selected user
    components/                  Navbar, EventCard, EventList, BookingButton, etc.
    pages/                      HomePage, EventPage, MyBookingsPage, AuditLogPage
    App.jsx                     Router and context provider
    main.jsx                    Entry point
    index.css                   All styles
  nginx.conf                    SPA fallback + API proxy to backend
  Dockerfile                    Multi-stage: Node build then Nginx

docker-compose.yml              Three services: db, backend, frontend
Makefile                        Setup, run, inspect, and cleanup commands
```

### Layer responsibilities

- **Handler**: Parses HTTP requests, calls the service layer, writes HTTP responses. No business logic, no SQL.
- **Service**: Orchestrates repository calls. Currently thin wrappers -- business rules (booking windows, per-user limits, waitlists) would be added here.
- **Repository**: Raw SQL queries, transaction management. No HTTP awareness. Interfaces enable mocking for unit tests.

---

## Assumptions

- **No authentication.** Users are pre-seeded and selected via a dropdown. In a production system, user identity would come from a JWT token, and the `user_id` in request bodies would be replaced by the authenticated user from the token. The repository query that checks `user_id` in the cancel WHERE clause would not need to change.

- **No event management.** Events are pre-seeded via migration. There is no admin UI to create, update, or delete events. This is out of scope.

- **No pagination.** List endpoints return all records. The dataset is small (9 users, 9 events). In production, cursor-based pagination would be added.

- **Timestamps are stored in UTC** (`TIMESTAMPTZ`) and displayed in the user's local timezone by the frontend.

- **Audit logs are immutable.** They are never updated or deleted. There is no `updated_at` column on the audit log table.

- **The `booked_count` field on events is denormalized.** The normalized approach would be `SELECT COUNT(*) FROM bookings WHERE event_id = ? AND status = 'active'` on every request. This is slow and unsafe under concurrency (COUNT can return stale values between transactions under READ COMMITTED isolation). The denormalized counter with an atomic UPDATE is both faster and correct. The CHECK constraint (`booked_count <= capacity`) is a database-level safety net.

---

## Design Decisions

**Why `POST /events/{id}/book` instead of `POST /bookings`:** The event is the primary resource being acted on. The URL makes it immediately clear which event is being booked. With `POST /bookings`, the event ID would be buried in the request body.

**Why `POST /bookings/{id}/cancel` instead of `DELETE /bookings/{id}`:** Cancellation is a state change (`active` to `cancelled`), not a deletion. The booking record is preserved for audit purposes. DELETE implies the resource no longer exists, which is not the case.

**Why `gorilla/mux` instead of `chi` or `gin`:** Path parameter extraction, method-based routing, and middleware support without pulling in a full framework. Standard `net/http` compatible handlers.

**Why `database/sql` + `lib/pq` instead of `sqlx` or an ORM:** Full control over queries and transactions. More boilerplate, but every SQL statement is visible and debuggable. Important for demonstrating understanding of the transaction model.

**Why PostgreSQL over MongoDB:** Booking with capacity constraints is inherently relational. Multi-document transactions in MongoDB are clunky for this use case. PostgreSQL gives us ACID transactions, row-level locking on UPDATE, CHECK constraints, and partial unique indexes -- all of which are load-bearing in this system.

**Why `fetch` instead of `axios` on the frontend:** Zero dependencies beyond React and React Router. The API client is a thin wrapper around the native fetch API. Keeps the frontend dependency footprint minimal.
