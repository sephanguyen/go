package dto

import "github.com/manabie-com/backend/internal/golibs/chatvendor/agora/entities"

type GetAgoraAppInfoResponse struct {
	Timestamp   uint64             `json:"timestamp"`
	Application string             `json:"application"`
	Action      string             `json:"action"`
	Duration    int                `json:"duration"`
	Entities    []entities.AppInfo `json:"entities"`
}
