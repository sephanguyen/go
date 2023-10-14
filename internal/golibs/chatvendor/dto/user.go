package dto

import "github.com/manabie-com/backend/internal/golibs/chatvendor/entities"

type GetUserRequest struct {
	VendorUserID string
}

type GetUserResponse struct {
	entities.User
}

type CreateUserRequest struct {
	UserID       string
	VendorUserID string
}

type CreateUserResponse struct {
	entities.User
}
