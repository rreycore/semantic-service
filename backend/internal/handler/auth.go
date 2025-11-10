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

var refreshTokenExpires time.Duration = 7 * 24 * time.Hour
var accessTokenExpires time.Duration = 15 * time.Minute

func setTokensCookie(w http.ResponseWriter, accessToken, refreshToken string) {
	accessCookie := http.Cookie{
		Name:     "jwt",
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(accessTokenExpires),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(refreshTokenExpires),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &accessCookie)
	http.SetCookie(w, &refreshCookie)
}

func clearTokensCookie(w http.ResponseWriter) {
	accessCookie := http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, &accessCookie)
	http.SetCookie(w, &refreshCookie)
}

func (h *handler) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	_, accessToken, refreshToken, err := h.service.Register(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			errorMessage := err.Error()
			return Register409JSONResponse{Error: &errorMessage}, nil
		}
		return nil, err
	}

	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		return nil, fmt.Errorf("response writer not found in context")
	}

	setTokensCookie(w, accessToken, refreshToken)
	w.WriteHeader(http.StatusCreated)

	return nil, nil
}

func (h *handler) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
	accessToken, refreshToken, err := h.service.Login(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return Login401Response{}, nil
		}
		return nil, err
	}

	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		return nil, fmt.Errorf("response writer not found in context")
	}

	setTokensCookie(w, accessToken, refreshToken)
	w.WriteHeader(http.StatusOK)

	return nil, nil
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

	newAccessToken, newRefreshToken, err := h.service.Refresh(ctx, cookie.Value)
	if err != nil {
		clearTokensCookie(w)
		return Refresh401Response{}, nil
	}

	setTokensCookie(w, newAccessToken, newRefreshToken)
	w.WriteHeader(http.StatusOK)

	return nil, nil
}

func (h *handler) Logout(ctx context.Context, request LogoutRequestObject) (LogoutResponseObject, error) {
	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		return nil, fmt.Errorf("response writer not found in context")
	}

	clearTokensCookie(w)
	w.WriteHeader(http.StatusNoContent)

	return nil, nil
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

	clearTokensCookie(w)
	w.WriteHeader(http.StatusNoContent)

	return nil, nil
}
