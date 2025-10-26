package handler

import (
	"backend/internal/service"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

var expires time.Duration = 7 * 24 * time.Hour

func (h *handler) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	_, accessToken, refreshToken, err := h.service.Register(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			errorMessage := err.Error()
			return Register409JSONResponse{Error: &errorMessage}, nil
		}
		return nil, err
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(expires),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	return Register201JSONResponse{
		Body: LoginResponse{
			AccessToken: &accessToken,
		},
		Headers: Register201ResponseHeaders{
			SetCookie: cookie.String(),
		},
	}, nil
}

func (h *handler) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
	accessToken, refreshToken, err := h.service.Login(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return Login401Response{}, nil
		}
		return nil, err
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(expires),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	return Login200JSONResponse{
		Body: LoginResponse{
			AccessToken: &accessToken,
		},
		Headers: Login200ResponseHeaders{
			SetCookie: cookie.String(),
		},
	}, nil
}

func (h *handler) Refresh(ctx context.Context, request RefreshRequestObject) (RefreshResponseObject, error) {
	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		return nil, fmt.Errorf("response writer not found in context")
	}
	r, ok := ctx.Value(requestKey).(*http.Request)
	if !ok {
		return nil, fmt.Errorf("request not found in context")
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return Refresh401Response{}, nil
	}

	accessToken, newRefreshToken, err := h.service.Refresh(ctx, cookie.Value)
	if err != nil {
		if errors.Is(err, service.ErrRefreshTokenNotFound) {
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})

			return Refresh401Response{}, nil
		}
		return nil, err
	}

	newCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		Expires:  time.Now().Add(expires),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	return Refresh200JSONResponse{
		Body: LoginResponse{
			AccessToken: &accessToken,
		},
		Headers: Refresh200ResponseHeaders{
			SetCookie: newCookie.String(),
		},
	}, nil
}

func (h *handler) Logout(ctx context.Context, request LogoutRequestObject) (LogoutResponseObject, error) {
	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		return nil, fmt.Errorf("response writer not found in context")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	return Logout204Response{}, nil
}

func (h *handler) FullLogout(ctx context.Context, request FullLogoutRequestObject) (FullLogoutResponseObject, error) {
	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		return nil, fmt.Errorf("response writer not found in context")
	}

	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	if err := h.service.FullLogout(ctx, userID); err != nil {
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	return FullLogout204Response{}, nil
}
