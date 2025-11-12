-- name: CreateDocument :one
-- Создает запись о новом документе для указанного пользователя.
-- Возвращает полную запись о новом документе.
INSERT INTO documents (user_id, filename)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserDocuments :many
-- Возвращает список всех документов для конкретного пользователя, включая количество необработанных и общее количество эмбеддингов.
SELECT
  d.id,
  d.user_id,
  d.filename,
  (
    SELECT COUNT(*) FROM chunks c WHERE c.document_id = d.id AND c.embedding IS NULL
  ) AS null_embeddings_count,
  (
    SELECT COUNT(*) FROM chunks c WHERE c.document_id = d.id
  ) AS total_embeddings_count
FROM documents d
WHERE d.user_id = $1;

-- name: GetUserDocumentByID :one
-- Находит конкретный документ по его ID.
-- ВАЖНО: также проверяет user_id для безопасности, чтобы пользователь не мог получить чужой документ.
SELECT
  d.id,
  d.user_id,
  d.filename,
  (
    SELECT COUNT(*) FROM chunks c WHERE c.document_id = d.id AND c.embedding IS NULL
  ) AS null_embeddings_count,
  (
    SELECT COUNT(*) FROM chunks c WHERE c.document_id = d.id
  ) AS total_embeddings_count
FROM documents d
WHERE d.id = $1 AND d.user_id = $2
LIMIT 1;

-- name: DeleteUserDocument :exec
-- Удаляет документ по его ID.
-- ВАЖНО: также проверяет user_id для безопасности.
DELETE FROM documents
WHERE id = $1 AND user_id = $2;
