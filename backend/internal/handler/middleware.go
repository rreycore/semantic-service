package handler

import (
	"context"
	"net/http"
)

type contextKey string

const (
	responseWriterKey contextKey = "responseWriter"
	requestKey        contextKey = "request"
)

func contextInjector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создаем новый контекст, обогащенный w и r.
		ctx := context.WithValue(r.Context(), responseWriterKey, w)
		ctx = context.WithValue(ctx, requestKey, r)

		// Вызываем следующий обработчик с новым контекстом.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
