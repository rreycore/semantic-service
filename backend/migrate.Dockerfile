FROM debian:12

WORKDIR /app

RUN apt update
RUN apt install -y --no-install-recommends curl ca-certificates
RUN rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh

COPY db/migrations .
COPY migrate.sh .

RUN chmod +x migrate.sh

ENTRYPOINT ["./migrate.sh"]

CMD ["up"]
