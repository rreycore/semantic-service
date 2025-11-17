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
	chunkOverlap = 50
)

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

		semanticSplitter := NewTextSplitter(chunkSize, chunkOverlap)
		chunks := semanticSplitter.SplitText(string(fileContent))
		s.log.Info().Int("chunks_count", len(chunks)).Msg("Текст разбит на чанки")

		for i, chunkText := range chunks {
			if _, err := repo.CreateChunk(ctx, userID, createdDoc.ID, filename, chunkText); err != nil {
				s.log.Err(err).Int("chunk_index", i).Msg("Ошибка сохранения чанка")
				return err
			}
		}

		s.log.Debug().Msg("Чанки загружены")

		doc = createdDoc
		chunksLen := int64(len(chunks))
		doc.TotalEmbeddings = chunksLen
		doc.NullEmbeddings = chunksLen

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

type TextSplitter struct {
	ChunkSize    int
	ChunkOverlap int
	Separators   []string
}

func NewTextSplitter(chunkSize, chunkOverlap int) *TextSplitter {
	if chunkOverlap >= chunkSize {
		panic("ChunkOverlap must be smaller than ChunkSize")
	}
	return &TextSplitter{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
		Separators:   []string{"\n\n", "\n", ". ", " ", ""},
	}
}

func (s *TextSplitter) splitTextWithSeparators(text string, separators []string) []string {
	var finalChunks []string

	if text == "" {
		return finalChunks
	}

	if utf8.RuneCountInString(text) <= s.ChunkSize {
		return []string{text}
	}

	if len(separators) == 0 {
		runes := []rune(text)
		for i := 0; i < len(runes); i += s.ChunkSize {
			end := i + s.ChunkSize
			if end > len(runes) {
				end = len(runes)
			}
			finalChunks = append(finalChunks, string(runes[i:end]))
		}
		return finalChunks
	}

	separator := separators[0]
	remainingSeparators := separators[1:]

	splits := strings.Split(text, separator)
	var goodSplits []string
	for i, split := range splits {
		if split != "" {
			if i < len(splits)-1 {
				goodSplits = append(goodSplits, split+separator)
			} else {
				goodSplits = append(goodSplits, split)
			}
		}
	}

	for _, split := range goodSplits {
		if utf8.RuneCountInString(split) > s.ChunkSize {
			finalChunks = append(finalChunks, s.splitTextWithSeparators(split, remainingSeparators)...)
		} else {
			finalChunks = append(finalChunks, split)
		}
	}
	return finalChunks
}

func (s *TextSplitter) mergeSplits(splits []string) []string {
	var chunks []string
	var currentChunk []string
	currentLength := 0

	for _, split := range splits {
		splitLength := utf8.RuneCountInString(split)

		if currentLength+splitLength > s.ChunkSize && len(currentChunk) > 0 {
			chunkText := strings.Join(currentChunk, "")
			chunks = append(chunks, chunkText)

			var overlap []string
			overlapLength := 0
			for i := len(currentChunk) - 1; i >= 0; i-- {
				part := currentChunk[i]
				partLength := utf8.RuneCountInString(part)
				if overlapLength+partLength > s.ChunkOverlap && len(overlap) > 0 {
					break
				}
				overlapLength += partLength
				overlap = append([]string{part}, overlap...)
			}
			currentChunk = overlap
			currentLength = overlapLength
		}

		currentChunk = append(currentChunk, split)
		currentLength += splitLength
	}

	if len(currentChunk) > 0 {
		chunkText := strings.Join(currentChunk, "")
		chunks = append(chunks, chunkText)
	}

	return chunks
}

func (s *TextSplitter) SplitText(text string) []string {
	splits := s.splitTextWithSeparators(text, s.Separators)
	return s.mergeSplits(splits)
}
