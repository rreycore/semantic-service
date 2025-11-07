set dotenv-load

default: timescaledb db-setup up-all


backend:
    cd backend && go run cmd/app/main.go

tidy:
    go mod tidy

up-all:
    docker compose up --build -d

generate:
    cd backend && go generate ./...

sqlc-generate:
    cd backend && go tool sqlc generate -f db/sqlc/sqlc.yaml

oapi-codegen:
    cd backend && go tool oapi-codegen -config codegen.yaml openapi.yaml

show-api-docs:
    docker run --rm -p 8080:8080 -v "{{justfile_directory()}}/openapi.yaml":/usr/share/nginx/html/openapi.yaml -e URL=openapi.yaml swaggerapi/swagger-ui

migrate-up:
    docker compose run --rm --build db-migration

migrate-down:
    docker compose run --rm --build db-migration down

pgai-setup:
    docker compose run --rm --build pgai-setup

db-setup: pgai-setup migrate-up

setup-vectorizer:
    docker compose run --rm --entrypoint "python -m pgai install -d postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB" vectorizer-worker

timescaledb:
    docker compose up timescaledb -d

vectorizer-worker:
    docker compose up vectorizer-worker -d

update-timescaledb:
    docker compose pull timescaledb

build-frontend:
    cd frontend && bun run build

run-frontend:
    cd frontend && bun run start

frontend: build-frontend run-frontend

frontend-dev:
    cd frontend && bun run dev

frontend-generate-api:
    cd frontend && bun run generate-api
