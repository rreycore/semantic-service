package repository

import (
	"backend/internal/domain"
	"backend/internal/repository/queries"
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, string, error)
	GetUserById(ctx context.Context, id int64) (*domain.User, string, error)
}

func userToDomain(u queries.User) *domain.User {
	return &domain.User{
		ID:    u.ID,
		Email: u.Email,
	}
}

func (p *postgres) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	u, err := p.q.CreateUser(ctx, queries.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}
	return userToDomain(u), nil
}

func (p *postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	u, err := p.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	return userToDomain(u), u.PasswordHash, nil
}

func (p *postgres) GetUserById(ctx context.Context, id int64) (*domain.User, string, error) {
	u, err := p.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	return userToDomain(u), u.PasswordHash, nil
}
