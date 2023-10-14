// Package p contains a Pub/Sub Cloud Function.
package p

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// PubSubMessage is the payload of a Pub/Sub event.
// See the documentation for more details:
// https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage
type PubSubMessage struct {
	Data []byte `json:"data"`
}

type OverridedDbFlag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type MessagePayload struct {
	Instance         string            `json:"instance"`
	Project          string            `json:"project"`
	InstanceType     string            `json:"instance_type"`
	OverridedDbFlags []OverridedDbFlag `json:"overrided_db_flags"`
}

func checkExistedFlag(flag *sqladmin.DatabaseFlags, OverridedDbFlags []OverridedDbFlag) bool {

	for _, f := range OverridedDbFlags {
		if f.Name == flag.Name {
			return true
		}
	}
	return false
}

// ProcessPubSub consumes and processes a Pub/Sub message.
func ProcessPubSub(ctx context.Context, m PubSubMessage) error {
	var psData MessagePayload
	err := json.Unmarshal(m.Data, &psData)
	if err != nil {
		log.Printf("Cannot unmarshall msg - error: %v", err)
		log.Println(err)
		return err
	}

	log.Printf("Change instance %s to type %s", psData.Instance, psData.InstanceType)
	log.Printf("%s", psData.OverridedDbFlags)

	// Create an http.Client that uses Application Default Credentials.
	hc, err := google.DefaultClient(ctx, sqladmin.CloudPlatformScope)
	if err != nil {
		log.Printf("Cannot create http.Client - error: %v", err)
		return err
	}

	// Create the Google Cloud SQL service.
	newOption := option.WithHTTPClient(hc)
	service, err := sqladmin.NewService(context.Background(), newOption)
	if err != nil {
		log.Printf("Cannot create sqladmin new service - error: %v", err)
		return err
	}

	instance, err := service.Instances.Get(psData.Project, psData.Instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cannot get instance - error: %v", err)
	}

	var newDatabaseFlags []*sqladmin.DatabaseFlags
	databaseFlags := instance.Settings.DatabaseFlags
	if len(psData.OverridedDbFlags) > 0 {

		for _, flag := range databaseFlags {
			if checkExistedFlag(flag, psData.OverridedDbFlags) {
				continue
			}
			newDatabaseFlags = append(newDatabaseFlags, flag)
		}

		for _, flag := range psData.OverridedDbFlags {
			newDatabaseFlags = append(newDatabaseFlags, &sqladmin.DatabaseFlags{
				Name:  flag.Name,
				Value: flag.Value,
			})
		}
		log.Printf("%#v\n", newDatabaseFlags)
		databaseFlags = newDatabaseFlags
	}
	// See more examples at:
	// https://cloud.google.com/sql/docs/sqlserver/admin-api/rest/v1beta4/instances/patch
	rb := &sqladmin.DatabaseInstance{
		Settings: &sqladmin.Settings{
			ActivationPolicy: "ALWAYS",
			Tier:             psData.InstanceType,
			DatabaseFlags:    databaseFlags,
		},
	}

	resp, err := service.Instances.Patch(psData.Project, psData.Instance, rb).Context(ctx).Do()
	if err != nil {
		log.Printf("Cannot patch the instance - error: %v", err)
		slackErr := postToSlack("warning", "*Saving cost*", "Having issue when updating instance type")
		log.Println(slackErr.Error())
		return err
	}
	log.Printf("%#v\n", resp)

	msg := fmt.Sprintf("Instance %s is being updated to type: %s", psData.Instance, psData.InstanceType)
	slackErr := postToSlack("good", "*Saving cost*", msg)
	if slackErr != nil {
		log.Println(slackErr.Error())
		return slackErr
	}
	return nil
}

func postToSlack(color string, title string, instances string) error {
	message := slack.WebhookMessage{
		Text: title,
		Attachments: []slack.Attachment{
			{
				Text:  instances,
				Color: color,
			},
		},
	}

	url := os.Getenv("SLACK_WEBHOOK")
	err := slack.PostWebhookContext(context.Background(), url, &message)
	if err != nil {
		return err
	}

	return nil
}
