FROM python:3.13-slim-bookworm

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/
RUN apt update
RUN apt install -y curl
RUN rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .

RUN ["uv", "sync", "--all-packages"]

ENTRYPOINT [ "uv", "run", "-m", "src.main" ]
