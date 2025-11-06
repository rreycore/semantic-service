set dotenv-path := "./backend/.env"

backend:
    cd backend && go run cmd/app/main.go

tidy:
    go mod tidy

generate:
    cd backend && go generate ./...

sqlc-generate:
    cd backend && go tool sqlc generate -f db/sqlc/sqlc.yaml

oapi-codegen:
    cd backend && go tool oapi-codegen -config codegen.yaml openapi.yaml

show-api-docs:
    docker run --rm -p 8080:8080 -v "{{justfile_directory()}}/openapi.yaml":/usr/share/nginx/html/openapi.yaml -e URL=openapi.yaml swaggerapi/swagger-ui

migrate-up:
    cd backend && go tool goose -dir db/migrations postgres "host=$POSTGRES_HOST port=$POSTGRES_PORT user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB sslmode=disable" up

migrate-down:
    cd backend && go tool goose -dir db/migrations postgres "host=$POSTGRES_HOST port=$POSTGRES_PORT user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB sslmode=disable" down

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
