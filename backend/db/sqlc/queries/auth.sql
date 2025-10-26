-- name: CreateRefreshToken :exec
-- Сохраняет хеш нового refresh токена.
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetRefreshByTokenHash :one
-- Находит активную сессию по хешу refresh токена.
SELECT *
FROM refresh_tokens
WHERE token_hash = $1 AND expires_at > NOW()
LIMIT 1;

-- name: DeleteRefreshToken :exec
-- Удаляет конкретную сессию.
DELETE FROM refresh_tokens
WHERE token_hash = $1;

-- name: DeleteAllUserRefreshTokens :exec
-- Удаляет все сессии пользователя.
DELETE FROM refresh_tokens
WHERE user_id = $1;
