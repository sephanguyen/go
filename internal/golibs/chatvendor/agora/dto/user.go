package dto

import "github.com/manabie-com/backend/internal/golibs/chatvendor/agora/entities"

// GET /users/{{username}}
type GetUserResponse struct {
	Path      string          `json:"path"`
	URI       string          `json:"uri"`
	Timestamp uint64          `json:"timestamp"`
	Count     int             `json:"count"`
	Action    string          `json:"action"`
	Duration  int             `json:"duration"`
	Modified  uint64          `json:"modified"`
	Entities  []entities.User `json:"entities"`
}

// POST /users
type CreateUserRequest struct {
	// This is Agora user_id
	UserName string `json:"username"`
	// This is Manabie user_id
	UserID   string `json:"nickname"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Action          string          `json:"action"`
	Application     string          `json:"application"`
	ApplicationName string          `json:"applicationName"`
	Organization    string          `json:"organization"`
	Path            string          `json:"path"`
	URI             string          `json:"uri"`
	Timestamp       uint64          `json:"timestamp"`
	Duration        int             `json:"duration"`
	Entities        []entities.User `json:"entities"`
}
