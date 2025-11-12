package repository

import (
	"backend/internal/domain"
	"backend/internal/repository/queries"
	"context"
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, userID int64, filename string) (*domain.Document, error)
	GetUserDocuments(ctx context.Context, userID int64) ([]domain.Document, error)
	GetUserDocumentByID(ctx context.Context, id, userID int64) (*domain.Document, error)
	DeleteUserDocument(ctx context.Context, id, userID int64) error
}

func documentToDomain(d queries.Document) *domain.Document {
	return &domain.Document{
		ID:       d.ID,
		UserID:   d.UserID,
		Filename: d.Filename,
	}
}

func documentRowToDomain(d queries.GetUserDocumentByIDRow) *domain.Document {
	return &domain.Document{
		ID:              d.ID,
		UserID:          d.UserID,
		Filename:        d.Filename,
		NullEmbeddings:  d.NullEmbeddingsCount,
		TotalEmbeddings: d.TotalEmbeddingsCount,
	}
}

func documentsSearchRowToDomain(d queries.GetUserDocumentsRow) *domain.Document {
	return &domain.Document{
		ID:              d.ID,
		UserID:          d.UserID,
		Filename:        d.Filename,
		NullEmbeddings:  d.NullEmbeddingsCount,
		TotalEmbeddings: d.TotalEmbeddingsCount,
	}
}

func (p *postgres) CreateDocument(ctx context.Context, userID int64, filename string) (*domain.Document, error) {
	d, err := p.q.CreateDocument(ctx, queries.CreateDocumentParams{
		UserID:   userID,
		Filename: filename,
	})
	if err != nil {
		return nil, err
	}
	return documentToDomain(d), nil
}

func (p *postgres) GetUserDocuments(ctx context.Context, userID int64) ([]domain.Document, error) {
	docs, err := p.q.GetUserDocuments(ctx, userID)
	if err != nil {
		p.log.Error().Err(err).Int64("userID", userID).Msg("DATABASE ERROR: Ошибка при получении документов")
		return nil, err
	}

	domainDocs := make([]domain.Document, len(docs))
	for i, d := range docs {
		domainDocs[i] = *documentsSearchRowToDomain(d)
	}

	return domainDocs, nil
}

func (p *postgres) GetUserDocumentByID(ctx context.Context, id, userID int64) (*domain.Document, error) {
	d, err := p.q.GetUserDocumentByID(ctx, queries.GetUserDocumentByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return documentRowToDomain(d), nil
}

func (p *postgres) DeleteUserDocument(ctx context.Context, id, userID int64) error {
	return p.q.DeleteUserDocument(ctx, queries.DeleteUserDocumentParams{
		ID:     id,
		UserID: userID,
	})
}
