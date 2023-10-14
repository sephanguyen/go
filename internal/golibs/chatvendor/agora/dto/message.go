package dto

// POST /messages/msg_recall
type RecallMessageRequest struct {
	MessageID string `json:"msg_id"`
	To        string `json:"to"`
	From      string `json:"from"`
	ChatType  string `json:"chat_type"`
	Force     bool   `json:"force"`
}

type RecallMessageResponse struct {
	Path            string `json:"path"`
	Action          string `json:"action"`
	Application     string `json:"application"`
	ApplicationName string `json:"applicationName"`
	Organization    string `json:"organization"`
	URI             string `json:"uri"`
	Timestamp       uint64 `json:"timestamp"`
	Duration        int    `json:"duration"`
	Data            struct {
		Recalled  bool   `json:"recalled"`
		ChatType  string `json:"chattype"`
		To        string `json:"to"`
		From      string `json:"from"`
		MessageID string `json:"msg_id"`
	}
}
