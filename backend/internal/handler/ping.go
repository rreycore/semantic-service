package handler

import (
	"context"
)

func (h *handler) Ping(ctx context.Context, request PingRequestObject) (PingResponseObject, error) {
	return Ping200TextResponse("pong"), nil
}
