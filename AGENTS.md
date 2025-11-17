## Dev environment tips

- Use `just backend` to run backend
- Use `just frontend-dev` to run frontend in dev mode
- Use `just embedding-service` to run embedding service
- Use `pnpm install --filter <package_name>` to add the package to your workspace so your IDE can see it.
- Check the name field inside each package's package.json to confirm the right name—skip the top-level one.

## Build/Lint/Test Commands

### Backend (Go)

- **Run**: `just backend` | **Lint**: `golangci-lint run` | **Generate**: `just generate`
- **Migrations**: `just migrate-up/down` | **Test**: `go test -v ./path/to/package` (no tests yet)

### Frontend (Next.js/TypeScript)

- **Dev**: `just frontend-dev` | **Build**: `just build-frontend` | **Lint**: `cd frontend && bun run lint`
- **Generate API**: `just frontend-generate-api` | **Test**: No framework configured

### Embedding Service (Python/FastAPI)

- **Run**: `just embedding-service` | **Test**: (No tests yet)

## Code Style Guidelines

### Go

- `gofmt` formatting, PascalCase exports, camelCase unexported
- Clean arch: handler→service→repository, dependency injection
- Return errors, structured logging (zerolog), Chi router + JWT

### TypeScript/React

- Next.js 16 + App Router, strict TS, Biome linting
- PascalCase components, interfaces for props, Tailwind CSS

### Python

- FastAPI async, type hints everywhere, HTTPException for errors

### General

- OpenAPI 3.0 contracts, goose migrations, Docker Compose dev
- No comments unless complex logic, environment config
