# backend.Dockerfile

# --- СТАДИЯ 1: Сборка (Builder) ---
FROM golang AS builder

# 1. Устанавливаем рабочую директорию. Это будет корень нашего Go-модуля.
WORKDIR /app/backend

# 2. ОПТИМИЗАЦИЯ КЭША: Копируем ТОЛЬКО файлы зависимостей.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 3. Устанавливаем инструменты. Этот слой тоже кэшируется.
RUN go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# 4. Возвращаемся в /app, чтобы правильно скопировать openapi.yaml
WORKDIR /app

# 5. Копируем ТОЛЬКО то, что нужно для сборки.
COPY backend/. ./backend/
COPY openapi.yaml ./

# 6. Снова заходим в рабочую директорию Go-модуля.
WORKDIR /app/backend

# 7. Генерируем код. Теперь ../../../openapi.yaml будет правильно указывать на /app/openapi.yaml
RUN go generate ./...

# 8. Обновляем зависимости и собираем проект.
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /app/main ./cmd/app

# --- СТАДИЯ 2: Финальный образ (Runner) ---
FROM gcr.io/distroless/base-debian12

COPY --from=builder /app/main /main
COPY --from=builder /app/backend/configs ./configs

ENTRYPOINT ["/main"]
CMD ["-cfg", "configs/prod"]
