package withus

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

func NotifySyncDataStatus(slackClient alert.SlackFactory, config configurations.Config, orgID string, orgName string, status string) error {
	now := time.Now().UTC().UnixMilli() / 1000
	payload := fmt.Sprintf(`{
	"channel": "%s",
	"username": "Withus Daily Checking",
	"attachments": [
		{
			"color": "%s",
			"author_name": "%s",
			"title": "TSV data file",
			"title_link": "%s",
			"fields": [%s],
			"footer": "Slack API",
			"ts": %d
		}
	]}`,
		config.WithUsConfig.SlackChannel,
		colorByStatus(status),
		authorName(orgName, config.Common.Environment),
		tsvDateLink(orgID, config.WithUsConfig.BucketName, config.Common.IdentityPlatformProject),
		fieldsByStatus(status),
		now)

	return slackClient.SendByte([]byte(payload))
}

func NotifyWithusSyncDataStatus(slackClient alert.SlackFactory, config configurations.Config, orgName string, status string, errs []InternalError) error {
	if len(errs) == 0 {
		return nil
	}

	now := time.Now().UTC().UnixMilli() / 1000
	payload := fmt.Sprintf(`{
	"channel": "%s",
	"username": "Withus Daily Checking",
	"attachments": [
		{
			"color": "%s",
			"author_name": "%s",
			"fields": [%s],
			"footer": "Slack API",
			"ts": %d
		}
	]}`,
		config.WithUsConfig.WithusChannel,
		colorByStatus(status),
		authorName(orgName, config.Common.Environment),
		fieldsByStatusAndErrors(status, errs),
		now)

	return slackClient.SendByte([]byte(payload))
}

func authorName(orgName string, env string) string {
	authorNameEnv := ""
	switch env {
	case "staging", "stag":
		authorNameEnv = "STAGING"
	case "uat":
		authorNameEnv = "UAT"
	case "prod":
		authorNameEnv = "PROD"
	}

	return fmt.Sprintf("[%s] %s", authorNameEnv, orgName)
}

func fieldsByStatus(status string) string {
	switch status {
	case constant.StatusSuccess:
		return `
				{
					"title": "Status",
					"value": "Success :pepe-baby:",
					"short": true
				}
			`
	case constant.StatusFailed:
		return `
				{
					"title": "Status",
					"value": "Failed :thinking:",
					"short": true
				}
			`
	}

	return ""
}

func fieldsByStatusAndErrors(status string, errors []InternalError) string {
	messages := ""
	if status == constant.StatusFailed {
		for _, err := range errors {
			messages = fmt.Sprintf("%s\n%s", fmt.Sprintf(
				`{
	"value": "*Row*:%d\n\t*Message*: %s\n\t*UserID*: %s",
	"short": false
},`, err.Index, err.Error(), err.UserID), messages)
		}
		return messages
	}

	return messages
}

func tsvDateLink(orgID string, bucketName string, project string) string {
	baseURL := "https://console.cloud.google.com/storage/browser/_details"
	t := time.Now()
	fileName := ""

	switch orgID {
	case fmt.Sprint(constants.ManagaraBase):
		fileName = "%2F" + fmt.Sprintf("withus/W2-D6L_users%s.tsv", DataFileNameSuffix(t))
	case fmt.Sprint(constants.ManagaraHighSchool):
		fileName = "%2F" + fmt.Sprintf("itee/N1-M1_users%s.tsv", DataFileNameSuffix(t))
	}
	return fmt.Sprintf("%s/%s/%s;tab=live_object?project=%s", baseURL, bucketName, fileName, project)
}

func colorByStatus(status string) string {
	switch status {
	case "success":
		return "#2EB886"
	case "failed":
		return "#A00003"
	default:
		return ""
	}
}
