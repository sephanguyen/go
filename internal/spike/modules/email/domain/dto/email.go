package dto

type Email struct {
	EmailID           string
	SendGridMessageID string
	Subject           string
	Content           EmailContent
	EmailFrom         string
	Status            string
	EmailRecipients   []string
}

type EmailContent struct {
	PlainTextContent string `json:"plain_text_content"`
	HTMLContent      string `json:"html_content"`
}
