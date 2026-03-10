# ============================================================
# Event Booking System — Makefile
# Two modes: local (manual Postgres + go run + vite dev)
#             docker (docker-compose)
# ============================================================

# --- Config ---
PG_CONTAINER   := postgres16
PG_PASSWORD    := pgsql
PG_PORT        := 5432
PG_DATA        := ~/postgres/data
DB_NAME        := eventbooking
DB_USER        := eventapp
DB_USER_PASS   := eventapp
DATABASE_URL   := postgres://$(DB_USER):$(DB_USER_PASS)@localhost:$(PG_PORT)/$(DB_NAME)?sslmode=disable
BACKEND_PORT   := 8080
FRONTEND_PORT  := 3000
LOG_OUTPUT     := ./event-booking.log
LOG_LEVEL      := info

# Shorthand: run psql inside the Postgres container
PSQL_ROOT      := docker exec -e PGPASSWORD=$(PG_PASSWORD) $(PG_CONTAINER) psql -U postgres
PSQL_APP       := docker exec -e PGPASSWORD=$(DB_USER_PASS) $(PG_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

# ============================================================
# LOCAL — Setup
# ============================================================

## Start a Postgres 16 container (persistent data at ~/postgres/data)
.PHONY: db-start
db-start:
	@echo "==> Starting Postgres container..."
	@docker run --name $(PG_CONTAINER) \
		-v $(PG_DATA):/var/lib/postgresql/data \
		-d -p $(PG_PORT):5432 \
		--restart unless-stopped \
		-e POSTGRES_PASSWORD=$(PG_PASSWORD) \
		postgres:16 || true
	@echo "==> Waiting for Postgres to be ready..."
	@until docker exec $(PG_CONTAINER) pg_isready -U postgres > /dev/null 2>&1; do sleep 1; done
	@echo "==> Postgres is ready."

## Create the app database and user
.PHONY: db-create
db-create:
	@echo "==> Verifying superuser connection..."
	@$(PSQL_ROOT) -c "SELECT 1" > /dev/null || \
		(echo "ERROR: Cannot connect as postgres. Run: make clean-local && make db-setup"; exit 1)
	@echo "==> Creating user $(DB_USER)..."
	@$(PSQL_ROOT) -c "CREATE USER $(DB_USER) WITH PASSWORD '$(DB_USER_PASS)';" 2>&1 | grep -v "already exists" || true
	@echo "==> Creating database $(DB_NAME)..."
	@$(PSQL_ROOT) -c "CREATE DATABASE $(DB_NAME) OWNER $(DB_USER);" 2>&1 | grep -v "already exists" || true
	@echo "==> Verifying $(DB_USER) can connect..."
	@$(PSQL_APP) -c "SELECT 1" > /dev/null || \
		(echo "ERROR: $(DB_USER) cannot connect to $(DB_NAME). Check Postgres logs."; exit 1)
	@echo "==> Done."

## Run SQL migrations against the local database
.PHONY: db-migrate
db-migrate:
	@echo "==> Running migrations..."
	@for f in backend/migrations/*.sql; do \
		echo "  applying $$f"; \
		docker cp "$$f" $(PG_CONTAINER):/tmp/migration.sql && \
		docker exec -e PGPASSWORD=$(DB_USER_PASS) $(PG_CONTAINER) \
			psql -U $(DB_USER) -d $(DB_NAME) -f /tmp/migration.sql -q || exit 1; \
	done
	@echo "==> Migrations complete."

## Full local DB setup: start container + create user/db + run migrations
.PHONY: db-setup
db-setup: db-start db-create db-migrate
	@echo "==> Database fully set up."

# ============================================================
# LOCAL — Run
# ============================================================

## Start the Go backend (runs on port 8080)
.PHONY: backend
backend:
	@echo "==> Starting backend on :$(BACKEND_PORT) (logs: $(LOG_OUTPUT))..."
	cd backend && DATABASE_URL="$(DATABASE_URL)" PORT="$(BACKEND_PORT)" \
		LOG_OUTPUT="$(LOG_OUTPUT)" LOG_LEVEL="$(LOG_LEVEL)" go run ./cmd/server

## Install frontend dependencies
.PHONY: fe-install
fe-install:
	@echo "==> Installing frontend dependencies..."
	cd frontend && npm install

## Start the Vite dev server (runs on port 3000, proxies /api to backend)
.PHONY: frontend
frontend:
	@echo "==> Starting frontend on :$(FRONTEND_PORT)..."
	cd frontend && npm run dev

## Full local setup: DB + frontend deps + instructions
.PHONY: setup
setup: db-setup fe-install
	@echo ""
	@echo "============================================"
	@echo " Setup complete! Run in two terminals:"
	@echo "   Terminal 1:  make backend"
	@echo "   Terminal 2:  make frontend"
	@echo ""
	@echo " Backend:  http://localhost:$(BACKEND_PORT)"
	@echo " Frontend: http://localhost:$(FRONTEND_PORT)"
	@echo "============================================"

# ============================================================
# DOCKER COMPOSE
# ============================================================

## Build and start all containers (Postgres + backend + frontend)
.PHONY: docker-up
docker-up:
	@echo "==> Starting Docker Compose stack..."
	docker-compose up -d --build
	@echo "==> Stack is running."
	@echo "    Frontend: http://localhost"
	@echo "    Backend:  http://localhost:8080"

## Stop all containers (data persists in volume)
.PHONY: docker-down
docker-down:
	docker-compose down

## Rebuild and restart a single service (usage: make docker-rebuild SVC=frontend)
.PHONY: docker-rebuild
docker-rebuild:
	docker-compose up -d --build $(SVC)

## Tail backend log file inside the Docker container
.PHONY: compose-logs
compose-logs:
	docker-compose exec backend tail -f /var/log/event-booking.log

## Tail local backend log file
.PHONY: logs
logs:
	@if [ -f "$(LOG_OUTPUT)" ]; then \
		echo "==> Tailing $(LOG_OUTPUT)..."; \
		tail -f $(LOG_OUTPUT); \
	else \
		echo "Log file $(LOG_OUTPUT) not found."; \
		echo "Start the backend first: make backend"; \
	fi

# ============================================================
# Inspect
# ============================================================

## Open a psql shell to the app database (inside container)
.PHONY: db-shell
db-shell:
	docker exec -it -e PGPASSWORD=$(DB_USER_PASS) $(PG_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

## Show all indexes in the public schema
.PHONY: db-indexes
db-indexes:
	@$(PSQL_APP) -c "SELECT indexname, indexdef FROM pg_indexes WHERE schemaname = 'public';"

## Show all tables
.PHONY: db-tables
db-tables:
	@$(PSQL_APP) -c "\dt"

## Reset local DB: drop + recreate + re-migrate
.PHONY: reset
reset: db-drop db-create db-migrate
	@echo "==> Database reset complete."

# ============================================================
# Cleanup — Local
# ============================================================

## Drop the app database and user (keeps Postgres container running)
.PHONY: db-drop
db-drop:
	@echo "==> Dropping database and user..."
	@$(PSQL_ROOT) -c "DROP DATABASE IF EXISTS $(DB_NAME);" 2>/dev/null || true
	@$(PSQL_ROOT) -c "DROP USER IF EXISTS $(DB_USER);" 2>/dev/null || true
	@echo "==> Done."

## Stop and remove the Postgres container (data persists in ~/postgres/data)
.PHONY: db-stop
db-stop:
	@echo "==> Stopping Postgres container..."
	@docker stop $(PG_CONTAINER) 2>/dev/null || true
	@docker rm $(PG_CONTAINER) 2>/dev/null || true
	@echo "==> Container removed."

## Nuke local setup: remove container + delete data + remove node_modules + log
.PHONY: clean-local
clean-local:
	@echo "==> Nuking local setup..."
	@docker stop $(PG_CONTAINER) 2>/dev/null || true
	@docker rm $(PG_CONTAINER) 2>/dev/null || true
	@rm -rf $(PG_DATA)
	@rm -rf frontend/node_modules frontend/dist
	@rm -f ./event-booking.log
	@echo "==> Local setup nuked. Run 'make setup' to rebuild from scratch."

# ============================================================
# Cleanup — Docker Compose
# ============================================================

## Nuke Docker Compose: stop containers + remove volumes + remove images
.PHONY: clean-docker
clean-docker:
	@echo "==> Nuking Docker Compose setup..."
	docker-compose down -v --rmi all --remove-orphans
	@echo "==> Docker Compose nuked. Run 'make docker-up' to rebuild from scratch."

# ============================================================
# Help
# ============================================================

.PHONY: help
help:
	@echo ""
	@echo "Event Booking System — Available Commands"
	@echo "=========================================="
	@echo ""
	@echo "  LOCAL SETUP"
	@echo "    make setup          Full setup: Postgres + DB + migrations + npm install"
	@echo "    make db-setup       Start Postgres + create DB + run migrations"
	@echo "    make db-start       Start Postgres container only"
	@echo "    make db-create      Create user and database"
	@echo "    make db-migrate     Run SQL migrations"
	@echo "    make fe-install     Install frontend npm dependencies"
	@echo ""
	@echo "  LOCAL RUN"
	@echo "    make backend        Start Go backend on :$(BACKEND_PORT) (logs to $(LOG_OUTPUT))"
	@echo "    make frontend       Start Vite dev server on :$(FRONTEND_PORT)"
	@echo "    make logs           Tail backend log file ($(LOG_OUTPUT))"
	@echo ""
	@echo "  DOCKER COMPOSE"
	@echo "    make docker-up      Build and start all containers"
	@echo "    make docker-down    Stop all containers"
	@echo "    make docker-rebuild SVC=frontend   Rebuild one service"
	@echo "    make compose-logs   Tail backend log file (/var/log/event-booking.log)"
	@echo ""
	@echo "  INSPECT"
	@echo "    make db-shell       Open psql shell to $(DB_NAME)"
	@echo "    make db-indexes     Show all indexes"
	@echo "    make db-tables      Show all tables"
	@echo "    make reset          Drop DB + recreate + re-migrate"
	@echo ""
	@echo "  CLEANUP"
	@echo "    make clean-local    Nuke local: container + data + node_modules + log"
	@echo "    make clean-docker   Nuke Docker: containers + volumes + images"
	@echo ""
