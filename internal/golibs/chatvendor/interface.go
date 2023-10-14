package chatvendor

import (
	"github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
)

// nolint
type ChatVendorClient interface {
	// Authentication
	GetAppToken() (string, error)
	GetUserToken(userID string) (string, uint64, error)
	GetAppKey() string

	// User Management
	GetUser(req *dto.GetUserRequest) (*dto.GetUserResponse, error)
	CreateUser(req *dto.CreateUserRequest) (*dto.CreateUserResponse, error)

	// Conversation Management
	CreateConversation(req *dto.CreateConversationRequest) (*dto.CreateConversationResponse, error)
	AddConversationMembers(req *dto.AddConversationMembersRequest) (*dto.AddConversationMembersResponse, error)
	RemoveConversationMembers(req *dto.RemoveConversationMembersRequest) (*dto.RemoveConversationMembersResponse, error)

	// Message Management
	DeleteMessage(req *dto.DeleteMessageRequest) (*dto.DeleteMessageResponse, error)
}
