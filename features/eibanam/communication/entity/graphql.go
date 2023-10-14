package entity

import (
	"time"

	"github.com/hasura/go-graphql-client"
)

type GraphqlNotificationQuery struct {
	InfoNotifications []struct {
		NotificationID    string         `graphql:"notification_id"`
		NotificationMsgID string         `graphql:"notification_msg_id"`
		SentAt            time.Time      `graphql:"sent_at"`
		ReceiverIDs       []string       `graphql:"receiver_ids"`
		Status            string         `graphql:"status"`
		Type              string         `graphql:"type"`
		TargetGroups      graphql.String `graphql:"target_groups"`
		CreatedAt         time.Time      `graphql:"created_at"`
		UpdatedAt         time.Time      `graphql:"updated_at"`
		EditorID          string         `graphql:"editor_id"`
		Event             string         `graphql:"event"`
		ScheduledAt       time.Time      `graphql:"scheduled_at"`
	} `graphql:"info_notifications(where: {notification_id: {_eq: $notification_id}})"`
}

type GraphqlNotification struct {
	InfoNotifications []struct {
		NotificationID    string    `json:"notification_id"`
		NotificationMsgID string    `json:"notification_msg_id"`
		SentAt            time.Time `json:"sent_at"`
		ReceiverIDs       []string  `json:"receiver_ids"`
		Status            string    `json:"status"`
		Type              string    `json:"type"`
		TargetGroups      struct {
			GradeFilter struct {
				Type   string  `json:"type"`
				Grades []int32 `json:"grades"`
			} `json:"grade_filter"`
			CourseFilter struct {
				Type      string   `json:"type"`
				CourseIDs []string `json:"course_ids"`
			} `json:"course_filter"`
			UserGroupFilter struct {
				UserGroup []string `json:"user_group"`
			} `json:"user_group_filter"`
		} `json:"target_groups"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		EditorID    string    `json:"editor_id"`
		Event       string    `json:"event"`
		ScheduledAt time.Time `json:"scheduled_at"`
	} `json:"info_notifications"`
}

type GraphqlNotificationMsgQuery struct {
	InfoNotificationMsgs []struct {
		Content           graphql.String `graphql:"content"`
		CreatedAt         time.Time      `graphql:"created_at"`
		MediaIDs          []string       `graphql:"media_ids"`
		NotificationMsgID string         `graphql:"notification_msg_id"`
		Title             string         `graphql:"title"`
		UpdatedAt         time.Time      `graphql:"updated_at"`
	} `graphql:"info_notification_msgs(where: {notification_msg_id: {_eq: $notification_msg_id}})"`
}

type GraphqlNotificationMsg struct {
	InfoNotificationMsgs []struct {
		Content struct {
			Raw         string `json:"raw"`
			RenderedURL string `json:"rendered_url"`
		} `json:"content"`
		CreatedAt         time.Time `json:"created_at"`
		MediaIDs          []string  `json:"media_ids"`
		NotificationMsgID string    `json:"notification_msg_id"`
		Title             string    `json:"title"`
		UpdatedAt         time.Time `json:"updated_at"`
	} `json:"info_notification_msgs"`
}
