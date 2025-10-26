package handler

import (
	"context"

	"github.com/go-chi/jwtauth/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *handler) GetUserProfile(ctx context.Context, request GetUserProfileRequestObject) (GetUserProfileResponseObject, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID := int64(claims["user_id"].(float64))

	user, err := h.service.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return GetUserProfile200JSONResponse{
		Id:    &user.ID,
		Email: (*openapi_types.Email)(&user.Email),
	}, nil
}
