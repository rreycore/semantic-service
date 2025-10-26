package service

import (
	"backend/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type UserService interface {
	GetUserProfile(ctx context.Context, userID int64) (*domain.User, error)
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (s *service) GetUserProfile(ctx context.Context, userID int64) (*domain.User, error) {
	user, _, err := s.repo.GetUserById(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}
