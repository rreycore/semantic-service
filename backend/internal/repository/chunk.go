package repository

import (
	"backend/internal/domain"
	"backend/internal/repository/queries"
	"context"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

type ChunkRepository interface {
	CreateChunk(ctx context.Context, userID, documentID int64, title, chunkText string) (*domain.Chunk, error)
	GetChunksByDocumentID(ctx context.Context, documentID, userID int64) ([]domain.Chunk, error)
	SearchUserChunks(ctx context.Context, userID int64, embedding []float32, limit int32) ([]domain.SearchResult, error)
	SearchChunksInDocument(ctx context.Context, userID, documentID int64, embedding []float32, limit int32) ([]domain.SearchResult, error)
}

func chunkToDomain(c queries.Chunk) *domain.Chunk {
	return &domain.Chunk{
		ID:         c.ID,
		UserID:     c.UserID,
		DocumentID: c.DocumentID,
		Text:       c.Text,
	}
}

func chunkRowToDomain(c queries.CreateChunkRow) *domain.Chunk {
	return &domain.Chunk{
		ID:         c.ID,
		UserID:     c.UserID,
		DocumentID: c.DocumentID,
		Text:       c.Text,
	}
}

func searchResultToDomain(c queries.SearchUserChunksRow) *domain.SearchResult {
	return &domain.SearchResult{
		ID:         c.ID,
		DocumentID: c.DocumentID,
		Title:      c.Title,
		Text:       c.Text,
		Distance:   c.Distance.(float64),
	}
}

func (p *postgres) CreateChunk(ctx context.Context, userID, documentID int64, title, text string) (*domain.Chunk, error) {
	c, err := p.q.CreateChunk(ctx, queries.CreateChunkParams{
		UserID:     userID,
		DocumentID: documentID,
		Title:      title,
		Text:       text,
	})
	if err != nil {
		return nil, err
	}
	return chunkRowToDomain(c), nil
}

func (p *postgres) GetChunksByDocumentID(ctx context.Context, documentID, userID int64) ([]domain.Chunk, error) {
	chunks, err := p.q.GetChunksByDocumentID(ctx, queries.GetChunksByDocumentIDParams{
		DocumentID: documentID,
		UserID:     userID,
	})
	if err != nil {
		return nil, err
	}

	domainChunks := make([]domain.Chunk, len(chunks))
	for i, c := range chunks {
		domainChunks[i] = *chunkToDomain(c)
	}

	return domainChunks, nil
}

func (p *postgres) SearchUserChunks(ctx context.Context, userID int64, embedding []float32, limit int32) ([]domain.SearchResult, error) {
	queryVector := pgvector.NewVector(embedding)
	results, err := p.q.SearchUserChunks(ctx, queries.SearchUserChunksParams{
		UserID:    userID,
		Embedding: queryVector,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	domainResults := make([]domain.SearchResult, len(results))
	for i, r := range results {
		_, ok := r.Distance.(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected type for distance: %T", r.Distance)
		}

		domainResults[i] = *searchResultToDomain(r)
	}

	return domainResults, nil
}

func (p *postgres) SearchChunksInDocument(ctx context.Context, userID, documentID int64, embedding []float32, limit int32) ([]domain.SearchResult, error) {
	queryVector := pgvector.NewVector(embedding)
	results, err := p.q.SearchChunksInDocument(ctx, queries.SearchChunksInDocumentParams{
		UserID:     userID,
		DocumentID: documentID,
		Embedding:  queryVector,
		Limit:      limit,
	})
	if err != nil {
		return nil, err
	}

	domainResults := make([]domain.SearchResult, len(results))
	for i, r := range results {
		distance, ok := r.Distance.(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected type for distance: %T", r.Distance)
		}

		domainResults[i] = domain.SearchResult{
			ID:         r.ID,
			DocumentID: r.DocumentID,
			Text:       r.Text,
			Distance:   distance,
		}
	}

	return domainResults, nil
}
