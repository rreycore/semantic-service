FROM golang as build-deps

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0

COPY . .
RUN go generate ./...
RUN go mod tidy

RUN CGO_ENABLED=0 go build cmd/app/main.go

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/main /main
COPY --from=builder /app/configs ./configs

ENTRYPOINT /main
