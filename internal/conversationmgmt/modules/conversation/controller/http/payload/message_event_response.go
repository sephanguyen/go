package payload

import "github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/constants"

type MessageEventResponse struct {
	Valid bool   `json:"valid"`
	Code  string `json:"code"`
}

func NewMessageEventResponse(isValid bool, err error) MessageEventResponse {
	if err != nil {
		return MessageEventResponse{
			Valid: isValid,
			Code:  constants.ResponseWebhookFAILED,
		}
	}

	return MessageEventResponse{
		Valid: isValid,
		Code:  constants.ResponseWebhookOK,
	}
}
