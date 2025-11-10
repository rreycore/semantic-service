default: timescaledb db-setup vectorizer-worker up-all

db-setup: vectorizer-setup migrate-up

backend:
    cd backend && go run cmd/app/main.go

tidy:
    go mod tidy

up-all:
    docker compose up --build -d

generate:
    cd backend && go generate ./...
    cd frontend && bun run generate-api

show-api-docs:
    docker run --rm -p 8080:8080 -v "{{justfile_directory()}}/openapi.yaml":/usr/share/nginx/html/openapi.yaml -e URL=openapi.yaml swaggerapi/swagger-ui

timescaledb:
    docker compose up timescaledb -d

migrate-up:
    docker compose run --rm --no-deps --build db-migration

migrate-down:
    docker compose run --rm --no-deps --build db-migration down

vectorizer-setup:
    docker compose --env-file .env.db --env-file .env.docker run --rm --no-deps vectorizer-setup

vectorizer-worker:
    docker compose --env-file .env.db --env-file .env.docker up vectorizer-worker -d

embedding-service:
    docker compose up embedding-service -d

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
