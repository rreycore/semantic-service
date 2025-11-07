package service

import (
	"backend/internal/domain"
	"backend/internal/embedding_client"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
)

type DocumentService interface {
	UploadAndProcessDocument(ctx context.Context, userID int64, filename string, fileContent []byte) (*domain.Document, error)
	SearchInDocument(ctx context.Context, userID, documentID int64, query string) ([]domain.SearchResult, error)
	ListUserDocuments(ctx context.Context, userID int64) ([]domain.Document, error)
	DeleteUserDocument(ctx context.Context, userID, documentID int64) error
	GetDocumentByID(ctx context.Context, userID, documentID int64) (*domain.Document, error)
}

var (
	ErrDocumentNotFound = errors.New("document not found or access denied")
)

const (
	chunkSize    = 200 // Количество слов в одном чанке
	chunkOverlap = 20  // Количество слов для перекрытия между чанками
	searchLimit  = 10  // Количество результатов при поиске
)

func (s *service) UploadAndProcessDocument(ctx context.Context, userID int64, filename string, fileContent []byte) (*domain.Document, error) {
	var doc *domain.Document

	err := s.repo.WithTransaction(ctx, func(repo repository.Repository) error {
		createdDoc, err := repo.CreateDocument(ctx, userID, filename)
		if err != nil {
			return err
		}

		chunks := splitIntoChunks(string(fileContent), chunkSize, chunkOverlap)

		for _, chunkText := range chunks {
			if _, err := repo.CreateChunk(ctx, userID, createdDoc.ID, chunkText); err != nil {
				return err
			}
		}

		doc = createdDoc
		return nil
	})

	return doc, err
}

func (s *service) SearchInDocument(ctx context.Context, userID, documentID int64, query string) ([]domain.SearchResult, error) {
	_, err := s.repo.GetUserDocumentByID(ctx, documentID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	embReq := embedding_client.EmbeddingRequest{Input: query}
	embResp, err := s.embeddingClient.CreateEmbeddings(ctx, embReq)
	if err != nil {
		s.log.Err(err).Msg("failed to get embedding for query")
		return nil, err
	}
	if len(embResp.Data) == 0 {
		return nil, errors.New("embedding service returned no embeddings for query")
	}

	embedding64 := embResp.Data[0].Embedding
	embedding32 := make([]float32, len(embedding64))
	for i, v := range embedding64 {
		embedding32[i] = float32(v)
	}
	queryVector := pgvector.NewVector(embedding32)

	return s.repo.SearchChunksInDocument(ctx, userID, documentID, queryVector, searchLimit)
}

func (s *service) ListUserDocuments(ctx context.Context, userID int64) ([]domain.Document, error) {
	return s.repo.GetUserDocuments(ctx, userID)
}

func (s *service) DeleteUserDocument(ctx context.Context, userID, documentID int64) error {
	return s.repo.DeleteUserDocument(ctx, documentID, userID)
}

// --- Вспомогательные функции ---

// splitIntoChunks - это простая реализация нарезки текста на чанки.
// В реальном проекте здесь может быть более сложная логика или внешняя библиотека.
func splitIntoChunks(text string, size, overlap int) []string {
	words := strings.Fields(text)
	var chunks []string
	if len(words) == 0 {
		return chunks
	}

	for i := 0; i < len(words); i += size - overlap {
		end := min(i+size, len(words))
		chunks = append(chunks, strings.Join(words[i:end], " "))
		if end == len(words) {
			break
		}
	}
	return chunks
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
