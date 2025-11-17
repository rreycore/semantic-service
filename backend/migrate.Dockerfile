FROM golang:1.23 AS builder

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM debian:12-slim

WORKDIR /app

COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY db/migrations/ .

COPY migrate.sh .

RUN chmod +x migrate.sh

ENTRYPOINT ["./migrate.sh"]
CMD ["up"]
