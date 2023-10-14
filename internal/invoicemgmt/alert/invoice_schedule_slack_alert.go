package alert

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/alert"
)

type Status string

const (
	Success Status = "success"
	Failed  Status = "failed"

	successColor      = "#2EB886"
	failColor         = "#A00003"
	attachmentTitle   = "Invoice Schedule Run Info"
	attachmentPreText = "A report on the recently completed invoice schedule job"
	footer            = "Slack API"
	userName          = "Invoice Schedule Checker Alert"
)

type Config struct {
	Environment  string
	SlackChannel string
}

type InvoiceScheduleConfig struct {
	SlackChannel string `yaml:"slack_channel"`
}

type InvoiceScheduleSlackAlert struct {
	slackClient alert.SlackFactory
	config      Config
}

func NewInvoiceScheduleSlackAlert(slackClient alert.SlackFactory, config Config) *InvoiceScheduleSlackAlert {
	return &InvoiceScheduleSlackAlert{
		slackClient: slackClient,
		config:      config,
	}
}

func (a *InvoiceScheduleSlackAlert) SendSuccessNotification() error {
	return a.sendNotification(Success, successFields())
}

func (a *InvoiceScheduleSlackAlert) SendFailNotification(err error) error {
	return a.sendNotification(Failed, failedFields(err))
}

func (a *InvoiceScheduleSlackAlert) sendNotification(s Status, fields []interface{}) error {
	if strings.TrimSpace(a.config.SlackChannel) == "" {
		return errors.New("slack channel is not provided")
	}

	payloadMap := map[string]interface{}{
		"channel":     a.config.SlackChannel,
		"username":    userName,
		"attachments": a.getAttachments(s, fields),
	}

	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return fmt.Errorf("error marshaling payload to JSON: %v", err)
	}

	return a.slackClient.SendByte(payloadJSON)
}

func getAuthorName(env string) string {
	authorNameEnv := ""
	switch env {
	case "local":
		authorNameEnv = "LOCAL"
	case "staging", "stag":
		authorNameEnv = "STAGING"
	case "uat":
		authorNameEnv = "UAT"
	case "prod":
		authorNameEnv = "PROD"
	}

	return fmt.Sprintf("[%s] Invoice Schedule Checker", authorNameEnv)
}

func successFields() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"title": "Status",
			"value": "Success :white_check_mark:",
			"short": true,
		},
		map[string]interface{}{
			"title": "Description",
			"value": "Invoice Schedule Job ran successfully :pepepoclap:",
			"short": true,
		},
	}
}

func failedFields(err error) []interface{} {
	fieldErr := ""
	if err != nil {
		fieldErr = err.Error()
	}

	return []interface{}{
		map[string]interface{}{
			"title": "Status",
			"value": "Failed :x:",
			"short": true,
		},
		map[string]interface{}{
			"title": "Description",
			"value": "Invoice Schedule Job encountered an error :pepe-bonk: \n Check the logs for more information",
			"short": true,
		},
		map[string]interface{}{
			"title": "Error Details",
			"value": fieldErr,
			"short": false,
		},
	}
}

func (a *InvoiceScheduleSlackAlert) getAttachments(s Status, fields []interface{}) []interface{} {
	var color string

	switch s {
	case Success:
		color = successColor
	case Failed:
		color = failColor
	}

	return []interface{}{
		map[string]interface{}{
			"color":       color,
			"author_name": getAuthorName(a.config.Environment),
			"title":       attachmentTitle,
			"pretext":     attachmentPreText,
			"fields":      fields,
			"footer":      footer,
			"ts":          time.Now().UTC().UnixMilli() / 1000,
		},
	}
}
