package handler

import (
	"backend/internal/config"
	"backend/internal/service"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog"
)

type MyHandler interface {
	StrictServerInterface
	Init() http.Handler
}

//go:generate just oapi-codegen
type handler struct {
	cfg       *config.HandlerConfig
	service   service.Service
	tokenAuth *jwtauth.JWTAuth
	log       *zerolog.Logger
}

func NewHandler(
	cfg *config.HandlerConfig,
	service service.Service,
	tokenAuth *jwtauth.JWTAuth,
	log *zerolog.Logger,
) MyHandler {
	return &handler{
		cfg:       cfg,
		service:   service,
		tokenAuth: tokenAuth,
		log:       log,
	}
}

func (h *handler) Init() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.AllowAll().Handler)

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	jwtMiddleware := []func(http.Handler) http.Handler{
		jwtauth.Verifier(h.tokenAuth),
		jwtauth.Authenticator(h.tokenAuth),
	}

	strictHandlerOptions := StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  h.requestErrorHandler,
		ResponseErrorHandlerFunc: h.responseErrorHandler,
	}

	strictHandler := NewStrictHandlerWithOptions(h, nil, strictHandlerOptions)
	wrapper := ServerInterfaceWrapper{
		Handler: strictHandler,
	}

	r.Get("/ping", wrapper.Ping)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", wrapper.Register)
		r.Post("/login", wrapper.Login)
		r.Post("/refresh", wrapper.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(jwtMiddleware...)

			r.Post("/logout", wrapper.Logout)
			r.Post("/full_logout", wrapper.FullLogout)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware...)

		r.Route("/users", func(r chi.Router) {
			r.Get("/me", wrapper.GetUserProfile)
		})

		r.Route("/documents", func(r chi.Router) {
			r.Post("/", wrapper.UploadDocument)
			r.Get("/", wrapper.ListUserDocuments)
			r.Get("/{documentID}", wrapper.GetDocumentByID)
			r.Delete("/{documentID}", wrapper.DeleteDocument)
			r.Post("/{documentID}/search", wrapper.SearchInDocument)
		})
	})

	return r
}

func (h *handler) requestErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error().Err(err).Msg("Request error")
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func (h *handler) responseErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	h.log.Error().Err(err).Str("uri", r.RequestURI).Msg("Internal server error")
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
