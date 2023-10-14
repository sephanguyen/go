package dto

// should only contain properties relate to our business domain
//
// if you need more data, just add more field here
type SGEmailEvent struct {
	OrganizationID       string `json:"organization_id"`
	EmailID              string `json:"email_id"`
	EmailRecipientID     string `json:"email_recipient_id"`
	Email                string `json:"email"`
	Event                string `json:"event"`
	SGEventID            string `json:"sg_event_id"`
	SGMessageID          string `json:"sg_message_id"`
	Reason               string `json:"reason"`
	Response             string `json:"response"`
	Status               string `json:"status"`
	Attempt              string `json:"attempt"`
	Type                 string `json:"type"`
	BounceClassification string `json:"bounce_classification"`
	Timestamp            int64  `json:"timestamp"`
}
