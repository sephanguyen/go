package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SlackFactory interface {
	Send(payload Payload) error
	SendByte(payload []byte) error
}

type IAttachment interface {
	AddField(field Field)
	AddSourceInfo(partnerName, env string)
	AddDetailInfo(missingNumber int, typ string)
}

type SlackImpl struct {
	WebHookURL string `json:"web_hook_url,omitempty"`
	HTTPClient http.Client
}

func (s *SlackImpl) Send(payload Payload) error {
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", s.WebHookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("unable to make new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("unable sending msg. Status: %v", resp.Status)
	}
	return nil
}

func (s *SlackImpl) SendByte(payload []byte) error {
	req, err := http.NewRequest("POST", s.WebHookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("unable to make new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("unable sending msg. Status: %v", resp.Status)
	}
	return nil
}

func InitAttachment(level string) *Attachment {
	var color string
	switch level {
	case "info":
		color = "#228B22"
	case "error":
		color = "#FF0000"
	case "warning":
		color = "#FF9900"
	}
	return &Attachment{
		Color:     color,
		Timestamp: time.Now().Unix(),
	}
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type Attachment struct {
	Color     string   `json:"color"`
	Text      *string  `json:"text"`
	Fields    []*Field `json:"fields"`
	Timestamp int64    `json:"ts"`
}

// slack payload
type Payload struct {
	Text        string        `json:"text,omitempty"`
	Attachments []IAttachment `json:"attachments,omitempty"`
}

func (att *Attachment) AddField(field Field) {
	att.Fields = append(att.Fields, &field)
}

func (att *Attachment) AddSourceInfo(partnerName, env string) {
	att.AddField(
		Field{
			Title: "Source info",
			Value: fmt.Sprintf("Partner: %s\nEnv: %s", partnerName, env),
		},
	)
}

func (att *Attachment) AddDetailInfo(missingNumber int, typ string) {
	att.AddField(
		Field{
			Title: "Detail",
			Value: fmt.Sprintf("Total missing items: %d\nType: %s", missingNumber, typ),
		},
	)
}

func (att *Attachment) AddErrorInfo(err error) {
	att.AddField(
		Field{
			Title: "Detail",
			Value: fmt.Sprintf("Date and time: %s\nError: %s", time.Now().Format("2006-01-02 15:04:05"), err.Error()),
		},
	)
}
