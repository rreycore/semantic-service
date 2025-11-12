FROM golang:1.23 AS builder

# Устанавливаем goose через go install
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# --- Финальный образ ---
FROM debian:12-slim

WORKDIR /app

# Копируем только goose binary из builder
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Копируем файлы миграций НАПРЯМУЮ в /app (не в подпапку)
COPY db/migrations/ .

# Копируем скрипт
COPY migrate.sh .

RUN chmod +x migrate.sh

ENTRYPOINT ["./migrate.sh"]
CMD ["up"]
