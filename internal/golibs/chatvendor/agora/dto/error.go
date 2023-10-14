package dto

type ErrorResponse struct {
	Exception        string `json:"exception"`
	Duration         int    `json:"duration"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Timestamp        uint64 `json:"timestamp"`
}
