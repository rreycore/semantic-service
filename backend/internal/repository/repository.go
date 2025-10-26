package repository

import (
	"backend/internal/repository/queries"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

//go:generate just sqlc-generate
type Repository interface {
	WithTransaction(ctx context.Context, fn func(repo Repository) error) error

	AuthRepository
	UserRepository
	DocumentRepository
	ChunkRepository
}

type postgres struct {
	pool *pgxpool.Pool
	q    *queries.Queries
	log  *zerolog.Logger
}

func NewPostgres(pool *pgxpool.Pool, log *zerolog.Logger) Repository {
	return &postgres{pool: pool, log: log}
}

func (p *postgres) WithTransaction(ctx context.Context, fn func(repo Repository) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer p.rollbackTx(tx)

	if err := fn(p.withTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *postgres) withTx(tx pgx.Tx) *postgres {
	return &postgres{
		q:    p.q.WithTx(tx),
		pool: p.pool,
		log:  p.log,
	}
}

func (p *postgres) rollbackTx(tx pgx.Tx) {
	err := tx.Rollback(context.Background())
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		p.log.Error().Err(err).Msg("RollbackTx")
	}
}

func calcOffsetLimit(page, size int64) (int32, int32) {
	return int32((page - 1) * size), int32(size)
}
