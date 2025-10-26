package repository

import (
	"backend/internal/domain"
	"backend/internal/repository/queries"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type AuthRepository interface {
	CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetRefreshByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteAllUserRefreshTokens(ctx context.Context, userID int64) error
}

func refreshTokenToDomain(t queries.RefreshToken) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt.Time,
	}
}

func (p *postgres) CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	return p.q.CreateRefreshToken(ctx, queries.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (p *postgres) GetRefreshByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	r, err := p.q.GetRefreshByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return refreshTokenToDomain(r), nil
}

func (p *postgres) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	return p.q.DeleteRefreshToken(ctx, tokenHash)
}

func (p *postgres) DeleteAllUserRefreshTokens(ctx context.Context, userID int64) error {
	return p.q.DeleteAllUserRefreshTokens(ctx, userID)
}
