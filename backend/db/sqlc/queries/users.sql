-- name: CreateUser :one
-- Создает нового пользователя.
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
-- Находит пользователя по email.
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
-- Находит пользователя по ID.
SELECT *
FROM users
WHERE id = $1
LIMIT 1;
