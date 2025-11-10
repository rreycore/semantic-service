-- name: CreateChunk :one
-- Создает один чанк для документа.
-- Поле 'embedding' здесь не передается, оно будет NULL при первичной вставке.
INSERT INTO chunks (user_id, document_id, title, text)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, document_id, title, text;

-- name: GetChunksByDocumentID :many
-- Возвращает все чанки для конкретного документа (для отображения или сборки полного текста).
-- ВАЖНО: также проверяет user_id для безопасности.
SELECT *
FROM chunks
WHERE document_id = $1 AND user_id = $2
ORDER BY id; -- Сортировка по ID, чтобы чанки шли в порядке их создания

-- name: SearchUserChunks :many
-- Самый важный запрос: выполняет семантический поиск по чанкам.
-- Находит N самых похожих чанков для заданного вектора-запроса, но только среди документов конкретного пользователя.
SELECT
    id,
    document_id,
    text,
    embedding <=> $1 AS distance -- Рассчитываем косинусное расстояние до вектора-запроса
FROM chunks
WHERE user_id = $2 -- ВАЖНО: строгая фильтрация по пользователю
ORDER BY distance ASC -- Сортируем по возрастанию расстояния (самые похожие - в начале)
LIMIT $3; -- Ограничиваем количество результатов

-- name: SearchChunksInDocument :many
-- Выполняет семантический поиск по чанкам ОДНОГО документа.
SELECT
    id,
    document_id,
    text,
    embedding <=> $1 AS distance
FROM chunks
WHERE user_id = $2 AND document_id = $3
ORDER BY distance ASC
LIMIT $4;
