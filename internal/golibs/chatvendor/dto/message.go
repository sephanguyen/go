package dto

type DeleteMessageRequest struct {
	ConversationID  string
	VendorUserID    string
	VendorMessageID string
}

type DeleteMessageResponse struct {
	IsSuccess bool
}
