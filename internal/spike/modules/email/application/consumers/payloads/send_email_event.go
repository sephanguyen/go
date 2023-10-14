package payloads

type SendEmailEvent struct {
	EmailID           string       `json:"email_id"`
	SendGridMessageID string       `json:"sg_message_id"`
	Subject           string       `json:"subject"`
	Content           EmailContent `json:"content"`
	EmailFrom         string       `json:"email_from"`
	Status            string       `json:"status"`
	EmailRecipients   []string     `json:"email_recipients"`
}

type EmailContent struct {
	PlainTextContent string `json:"plain_text_content"`
	HTMLContent      string `json:"html_content"`
}
