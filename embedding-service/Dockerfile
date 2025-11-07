FROM python:3.13-slim-bookworm

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

RUN apt update
RUN apt install -y curl
RUN rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY pyproject.toml uv.lock ./
RUN pip install uv && uv sync --frozen --no-install-project

COPY . .

RUN uv sync --all-packages

ENTRYPOINT ["uv", "run", "fastapi", "run", "src/main.py", "--host", "0.0.0.0", "--port", "8000"]
