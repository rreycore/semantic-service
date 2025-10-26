-- name: CreateDocument :one
-- Создает запись о новом документе для указанного пользователя.
-- Возвращает полную запись о новом документе.
INSERT INTO documents (user_id, filename)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserDocuments :many
-- Возвращает список всех документов для конкретного пользователя.
SELECT *
FROM documents
WHERE user_id = $1
ORDER BY filename;

-- name: GetUserDocumentByID :one
-- Находит конкретный документ по его ID.
-- ВАЖНО: также проверяет user_id для безопасности, чтобы пользователь не мог получить чужой документ.
SELECT *
FROM documents
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: DeleteUserDocument :exec
-- Удаляет документ по его ID.
-- ВАЖНО: также проверяет user_id для безопасности.
DELETE FROM documents
WHERE id = $1 AND user_id = $2;
