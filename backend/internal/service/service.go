package service

import (
	"backend/internal/embedding_client"
	"backend/internal/repository"

	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog"
)

type Service interface {
	AuthService
	UserService
	DocumentService
}

type service struct {
	repo            repository.Repository
	tokenAuth       *jwtauth.JWTAuth
	embeddingClient *embedding_client.Client
	log             *zerolog.Logger
}

func New(
	repo repository.Repository,
	tokenAuth *jwtauth.JWTAuth,
	embeddingClient *embedding_client.Client,
	log *zerolog.Logger,
) Service {
	return &service{
		repo:            repo,
		tokenAuth:       tokenAuth,
		embeddingClient: embeddingClient,
		log:             log,
	}
}
