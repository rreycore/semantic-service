package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.User, string, string, error)
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	Refresh(ctx context.Context, token string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, token string) error
	FullLogout(ctx context.Context, userID int64) error
}

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

var (
	ErrUserExists           = errors.New("user with this email already exists")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrRefreshTokenNotFound = errors.New("refresh token not found or expired")
)

func (s *service) Register(ctx context.Context, email, password string) (*domain.User, string, string, error) {
	var user *domain.User
	var accessToken, refreshToken string

	err := s.repo.WithTransaction(ctx, func(repo repository.Repository) error {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, _, err = repo.GetUserByEmail(ctx, email)
		if err == nil {
			return ErrUserExists
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		createdUser, err := repo.CreateUser(ctx, email, string(passwordHash))
		if err != nil {
			return err
		}
		user = createdUser

		newAccessToken, newRefreshToken, err := s.generateTokenPair(ctx, repo, user)
		if err != nil {
			return err
		}
		accessToken = newAccessToken
		refreshToken = newRefreshToken

		return nil
	})

	return user, accessToken, refreshToken, err
}

func (s *service) Login(ctx context.Context, email, password string) (string, string, error) {
	var accessToken, refreshToken string

	err := s.repo.WithTransaction(ctx, func(repo repository.Repository) error {
		user, passwordHash, err := repo.GetUserByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidCredentials
			}
			return err
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
			return ErrInvalidCredentials
		}

		newAccessToken, newRefreshToken, err := s.generateTokenPair(ctx, repo, user)
		if err != nil {
			return err
		}

		accessToken = newAccessToken
		refreshToken = newRefreshToken
		return nil
	})

	return accessToken, refreshToken, err
}

func (s *service) Refresh(ctx context.Context, token string) (string, string, error) {
	var accessToken, newRefreshToken string

	err := s.repo.WithTransaction(ctx, func(repo repository.Repository) error {
		tokenHash := hashToken(token)

		refreshToken, err := repo.GetRefreshByTokenHash(ctx, tokenHash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRefreshTokenNotFound
			}
			return err
		}

		if err := repo.DeleteRefreshToken(ctx, tokenHash); err != nil {
			return err
		}

		user := &domain.User{ID: refreshToken.UserID}
		newAccessToken, generatedRefreshToken, err := s.generateTokenPair(ctx, repo, user)
		if err != nil {
			return err
		}

		accessToken = newAccessToken
		newRefreshToken = generatedRefreshToken
		return nil
	})

	return accessToken, newRefreshToken, err
}

func (s *service) Logout(ctx context.Context, token string) error {
	return s.repo.WithTransaction(ctx, func(repo repository.Repository) error {
		tokenHash := hashToken(token)
		return repo.DeleteRefreshToken(ctx, tokenHash)
	})
}

func (s *service) FullLogout(ctx context.Context, userID int64) error {
	return s.repo.WithTransaction(ctx, func(repo repository.Repository) error {
		return repo.DeleteAllUserRefreshTokens(ctx, userID)
	})
}

func (s *service) generateTokenPair(ctx context.Context, repo repository.Repository, user *domain.User) (string, string, error) {
	claims := map[string]interface{}{
		"user_id": user.ID,
		"exp":     jwtauth.ExpireIn(accessTokenTTL),
		"iat":     time.Now().Unix(),
	}
	_, accessToken, err := s.tokenAuth.Encode(claims)
	if err != nil {
		return "", "", err
	}

	rawRefreshToken, err := generateSecureRandomString(32)
	if err != nil {
		return "", "", err
	}
	refreshTokenHash := hashToken(rawRefreshToken)
	expiresAt := time.Now().Add(refreshTokenTTL)

	if err := repo.CreateRefreshToken(ctx, user.ID, refreshTokenHash, expiresAt); err != nil {
		return "", "", err
	}

	return accessToken, rawRefreshToken, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func generateSecureRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
