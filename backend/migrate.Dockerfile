FROM debian:12

WORKDIR /app

RUN apt update
RUN apt install -y --no-install-recommends curl ca-certificates
RUN rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh

COPY db/migrations ./migrations

ENTRYPOINT goose postgres "host=$POSTGRES_HOST port=$POSTGRES_PORT user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DBNAME sslmode=$POSTGRES_SSLMODE" up
