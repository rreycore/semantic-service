package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

type DocumentService interface {
	UploadDocument(ctx context.Context, userID int64, filename string, fileContent []byte) (*domain.Document, error)
	Search(ctx context.Context, userID int64, query string) ([]domain.SearchResult, error)
	SearchInDocument(ctx context.Context, userID, documentID int64, query string) ([]domain.SearchResult, error)
	ListUserDocuments(ctx context.Context, userID int64) ([]domain.Document, error)
	DeleteUserDocument(ctx context.Context, userID, documentID int64) error
	GetDocumentByID(ctx context.Context, userID, documentID int64) (*domain.Document, error)
}

var (
	ErrDocumentNotFound = errors.New("document not found or access denied")
)

const (
	searchLimit  = 10
	chunkSize    = 1000
	chunkOverlap = 100
)

var separators = []string{"\n\n", "\n", ". ", " ", ""}

func (s *service) UploadDocument(ctx context.Context, userID int64, filename string, fileContent []byte) (*domain.Document, error) {
	s.log.Info().Int64("user_id", userID).Str("filename", filename).Int("size_bytes", len(fileContent)).Msg("Начало загрузки документа")

	var doc *domain.Document
	err := s.repo.WithTransaction(ctx, func(repo repository.Repository) error {
		createdDoc, err := repo.CreateDocument(ctx, userID, filename)
		if err != nil {
			s.log.Err(err).Msg("Ошибка создания документа в БД")
			return err
		}

		s.log.Info().Int64("doc_id", createdDoc.ID).Msg("Документ создан, начинаем чанкование")

		chunks := SplitTextIterative(string(fileContent), chunkSize, chunkOverlap)
		s.log.Info().Int("chunks_count", len(chunks)).Msg("Текст разбит на чанки")

		for i, chunkText := range chunks {
			if _, err := repo.CreateChunk(ctx, userID, createdDoc.ID, filename, chunkText); err != nil {
				s.log.Err(err).Int("chunk_index", i).Msg("Ошибка сохранения чанка")
				return err
			}
		}

		s.log.Debug().Msg("Чанки загружены")

		doc = createdDoc
		return nil
	})

	if err != nil {
		s.log.Err(err).Msg("Транзакция не удалась")
		return nil, err
	}

	s.log.Info().Int64("doc_id", doc.ID).Msg("Документ успешно загружен и чанки обработаны")
	return doc, nil
}

func (s *service) Search(ctx context.Context, userID int64, query string) ([]domain.SearchResult, error) {
	searchEmbedding, err := s.embeddingClient.CreateSearchEmbedding(ctx, query)
	if err != nil {
		s.log.Err(err).Msg("failed to get embedding for query")
		return nil, err
	}
	if len(searchEmbedding.Data) == 0 {
		return nil, errors.New("embedding service returned no embeddings for query")
	}

	embedding := searchEmbedding.Data[0].Embedding
	return s.repo.SearchUserChunks(ctx, userID, embedding, searchLimit)
}
func (s *service) SearchInDocument(ctx context.Context, userID, documentID int64, query string) ([]domain.SearchResult, error) {
	_, err := s.repo.GetUserDocumentByID(ctx, documentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	searchEmbedding, err := s.embeddingClient.CreateSearchEmbedding(ctx, query)
	if err != nil {
		s.log.Err(err).Msg("failed to get embedding for query")
		return nil, err
	}
	if len(searchEmbedding.Data) == 0 {
		return nil, errors.New("embedding service returned no embeddings for query")
	}

	embedding := searchEmbedding.Data[0].Embedding
	return s.repo.SearchChunksInDocument(ctx, userID, documentID, embedding, searchLimit)
}

func (s *service) ListUserDocuments(ctx context.Context, userID int64) ([]domain.Document, error) {
	return s.repo.GetUserDocuments(ctx, userID)
}

func (s *service) DeleteUserDocument(ctx context.Context, userID, documentID int64) error {
	return s.repo.DeleteUserDocument(ctx, documentID, userID)
}

func (s *service) GetDocumentByID(ctx context.Context, userID, documentID int64) (*domain.Document, error) {
	doc, err := s.repo.GetUserDocumentByID(ctx, documentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	return doc, nil
}

// === ИТЕРАТИВНЫЙ АЛГОРИТМ ДЕЛЕНИЯ ТЕКСТА ===

type splitTask struct {
	text     string
	sepIndex int // какой сепаратор использовать
}

func SplitTextIterative(text string, chunkSize, chunkOverlap int) []string {
	if utf8.RuneCountInString(text) <= chunkSize {
		return []string{text}
	}

	var result []string
	queue := []splitTask{{text: text, sepIndex: 0}}

	for len(queue) > 0 {
		task := queue[0]
		queue = queue[1:]

		// Если уже на последнем сепараторе — делим по размеру
		if task.sepIndex >= len(separators)-1 {
			chunks := splitBySizeWithOverlap(task.text, chunkSize, chunkOverlap)
			result = append(result, chunks...)
			continue
		}

		sep := separators[task.sepIndex]
		parts := strings.Split(task.text, sep)

		var current strings.Builder
		var tempChunks []string

		for i, part := range parts {
			addLen := len(part)
			if i > 0 {
				addLen += len(sep)
			}

			// Если не влезает — сохраняем текущий чанк
			if current.Len()+addLen > chunkSize && current.Len() > 0 {
				chunk := current.String()
				tempChunks = append(tempChunks, chunk)
				current.Reset()

				// Добавляем перекрытие
				if chunkOverlap > 0 {
					overlap := getLastNRunes(chunk, chunkOverlap)
					current.WriteString(overlap)
				}
			}

			if i > 0 && current.Len() > 0 {
				current.WriteString(sep)
			}
			current.WriteString(part)
		}

		if current.Len() > 0 {
			tempChunks = append(tempChunks, current.String())
		}

		// Обрабатываем каждый полученный чанк
		for _, chunk := range tempChunks {
			if utf8.RuneCountInString(chunk) > chunkSize {
				// Отправляем в очередь с следующим сепаратором
				queue = append(queue, splitTask{text: chunk, sepIndex: task.sepIndex + 1})
			} else {
				result = append(result, chunk)
			}
		}
	}

	return result
}

// splitBySizeWithOverlap — финальное деление по размеру
func splitBySizeWithOverlap(text string, chunkSize, chunkOverlap int) []string {
	runes := []rune(text)
	var chunks []string
	i := 0
	for i < len(runes) {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
		i = end - chunkOverlap
		if i <= 0 {
			i = end
		}
	}
	return chunks
}

// getLastNRunes — безопасно берёт последние N рун
func getLastNRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[len(runes)-n:])
}
