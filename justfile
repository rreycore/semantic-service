# Default: build and start everything
default:
    @echo "🔧 Running setup tasks..."
    docker compose --profile tools run --rm vectorizer-setup
    docker compose --profile tools run --rm --build db-migration
    @echo "🚀 Starting all services..."
    docker compose up --build -d
    @echo "✅ All services started!"
    @docker compose ps

# === Dev commands (local run without Docker) ===

backend:
    cd backend && go run cmd/app/main.go

tidy:
    go mod tidy

frontend-dev:
    cd frontend && bun run dev

# === Docker: Full environment ===

down:
    docker compose down

# === Docker: Build images ===

build:
    docker compose build

build-backend:
    docker compose build backend

build-frontend:
    docker compose build frontend

build-embedding:
    docker compose build embedding-service

# === Docker: Quick rebuild + recreate (safe for all changes) ===

rebuild-backend: build-backend
    docker compose up -d backend

rebuild-frontend: build-frontend
    docker compose up -d frontend

rebuild-embedding: build-embedding
    docker compose up -d embedding-service

# === Docker: Restart services (no rebuild) ===

restart-backend:
    docker compose restart backend

restart-frontend:
    docker compose restart frontend

restart-embedding:
    docker compose restart embedding-service

# === Docker: Individual services ===

timescaledb:
    docker compose up -d timescaledb

embedding-service:
    docker compose up --build -d embedding-service

vectorizer-worker:
    docker compose up -d vectorizer-worker

# === Database migrations ===

migrate-up:
    docker compose run --rm --build db-migration

migrate-down:
    docker compose run --rm --build db-migration down

vectorizer-setup:
    docker compose run --rm vectorizer-setup

# === Maintenance ===

update-timescaledb:
    docker compose pull timescaledb

# === Code generation ===

generate:
    cd backend && go generate ./...
    cd frontend && bun run generate-api

frontend-generate-api:
    cd frontend && bun run generate-api

# === Utilities ===

logs service:
    docker compose logs -f {{service}}

show-api-docs:
    docker run --rm -p 8080:8080 -v "{{justfile_directory()}}/openapi.yaml":/usr/share/nginx/html/openapi.yaml -e URL=openapi.yaml swaggerapi/swagger-ui

ping:
    echo "pong"
