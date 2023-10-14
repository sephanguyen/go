package entities

import (
	"encoding/json"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type InfoNotification struct {
	NotificationID             pgtype.Text
	NotificationMsgID          pgtype.Text
	Type                       pgtype.Text
	Data                       pgtype.JSONB
	EditorID                   pgtype.Text
	TargetGroups               pgtype.JSONB
	ReceiverIDs                pgtype.TextArray
	Event                      pgtype.Text
	Status                     pgtype.Text
	ScheduledAt                pgtype.Timestamptz
	Owner                      pgtype.Int4
	IsImportant                pgtype.Bool
	QuestionnaireID            pgtype.Text
	CreatedAt                  pgtype.Timestamptz
	UpdatedAt                  pgtype.Timestamptz
	DeletedAt                  pgtype.Timestamptz
	SentAt                     pgtype.Timestamptz
	CreatedUserID              pgtype.Text
	ExcludedGenericReceiverIDs pgtype.TextArray
	GenericReceiverIDs         pgtype.TextArray
	ReceiverNames              pgtype.TextArray
}

func (e *InfoNotification) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"notification_id",
		"notification_msg_id",
		"type",
		"data",
		"editor_id",
		"target_groups",
		"receiver_ids",
		"event",
		"status",
		"scheduled_at",
		"owner",
		"is_important",
		"questionnaire_id",
		"created_at",
		"updated_at",
		"deleted_at",
		"sent_at",
		"created_user_id",
		"excluded_generic_receiver_ids",
		"generic_receiver_ids",
		"receiver_names",
	}
	values = []interface{}{
		&e.NotificationID,
		&e.NotificationMsgID,
		&e.Type,
		&e.Data,
		&e.EditorID,
		&e.TargetGroups,
		&e.ReceiverIDs,
		&e.Event,
		&e.Status,
		&e.ScheduledAt,
		&e.Owner,
		&e.IsImportant,
		&e.QuestionnaireID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
		&e.SentAt,
		&e.CreatedUserID,
		&e.ExcludedGenericReceiverIDs,
		&e.GenericReceiverIDs,
		&e.ReceiverNames,
	}
	return
}

func (e *InfoNotification) GetTargetGroup() (*InfoNotificationTarget, error) {
	targetGroup := &InfoNotificationTarget{}
	err := e.TargetGroups.AssignTo(targetGroup)
	return targetGroup, err
}

func (*InfoNotification) TableName() string {
	return "info_notifications"
}

type InfoNotifications []*InfoNotification

// Add append new InfoNotification
func (u *InfoNotifications) Add() database.Entity {
	e := &InfoNotification{}
	*u = append(*u, e)

	return e
}

func (e *InfoNotification) IsMuteMode() bool {
	var data map[string]string
	err := json.Unmarshal(e.Data.Bytes, &data)
	if err != nil {
		data = make(map[string]string)
	}
	if mode, has := data["mode"]; has && mode == "mute" {
		return true
	}
	return false
}

type InfoNotificationMsg struct {
	NotificationMsgID pgtype.Text
	Title             pgtype.Text
	Content           pgtype.JSONB
	MediaIDs          pgtype.TextArray
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (e *InfoNotificationMsg) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"notification_msg_id",
		"title",
		"content",
		"media_ids",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.NotificationMsgID,
		&e.Title,
		&e.Content,
		&e.MediaIDs,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*InfoNotificationMsg) TableName() string {
	return "info_notification_msgs"
}

func (e *InfoNotificationMsg) GetContent() (*RichText, error) {
	content := &RichText{}
	err := e.Content.AssignTo(content)
	return content, err
}

type InfoNotificationMsgs []*InfoNotificationMsg

// Add append new InfoNotificationMsg
func (u *InfoNotificationMsgs) Add() database.Entity {
	e := &InfoNotificationMsg{}
	*u = append(*u, e)

	return e
}

type UserInfoNotification struct {
	UserNotificationID       pgtype.Text
	NotificationID           pgtype.Text
	UserID                   pgtype.Text
	Status                   pgtype.Text
	Courses                  pgtype.TextArray
	CurrentGrade             pgtype.Int2
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
	DeletedAt                pgtype.Timestamptz
	IsIndividual             pgtype.Bool
	UserGroup                pgtype.Text
	ParentID                 pgtype.Text
	StudentID                pgtype.Text
	QuestionnaireStatus      pgtype.Text
	QuestionnaireSubmittedAt pgtype.Timestamptz
	GradeID                  pgtype.Text
	ParentName               pgtype.Text
	StudentName              pgtype.Text
}

func (e *UserInfoNotification) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_notification_id",
		"notification_id",
		"user_id",
		"status",
		"course_ids",
		"current_grade",
		"created_at",
		"updated_at",
		"deleted_at",
		"is_individual",
		"user_group",
		"parent_id",
		"student_id",
		"qn_status",
		"qn_submitted_at",
		"grade_id",
		"parent_name",
		"student_name",
	}
	values = []interface{}{
		&e.UserNotificationID,
		&e.NotificationID,
		&e.UserID,
		&e.Status,
		&e.Courses,
		&e.CurrentGrade,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
		&e.IsIndividual,
		&e.UserGroup,
		&e.ParentID,
		&e.StudentID,
		&e.QuestionnaireStatus,
		&e.QuestionnaireSubmittedAt,
		&e.GradeID,
		&e.ParentName,
		&e.StudentName,
	}
	return
}

func (*UserInfoNotification) TableName() string {
	return "users_info_notifications"
}

type UserInfoNotifications []*UserInfoNotification

// Add append new InfoNotificationMsg
func (u *UserInfoNotifications) Add() database.Entity {
	e := &UserInfoNotification{}
	*u = append(*u, e)

	return e
}

// nolint
type InfoNotificationTarget_CourseFilter_Course struct {
	CourseID   string `json:"course_id"`
	CourseName string `json:"course_name"`
}

// nolint
type InfoNotificationTarget_CourseFilter struct {
	CourseIDs []string `json:"course_ids"`
	Type      string   `json:"type"`
	// This field only need when upserting a notification from Backoffice
	Courses []InfoNotificationTarget_CourseFilter_Course `json:"courses"`
}

func (filter *InfoNotificationTarget_CourseFilter) GetNameValues() []string {
	names := []string{}
	for _, course := range filter.Courses {
		names = append(names, course.CourseName)
	}
	return names
}

// nolint
type InfoNotificationTarget_GradeFilter struct {
	// Dereplicate: no longer support, please using Grade Master (switch to GradeIDs)
	Grades []int32 `json:"grades"`

	Type     string   `json:"type"`
	GradeIDs []string `json:"grade_ids"`
}

// nolint
type InfoNotificationTarget_LocationFilter_Location struct {
	LocationID   string `json:"location_id"`
	LocationName string `json:"location_name"`
}

// nolint
type InfoNotificationTarget_LocationFilter struct {
	LocationIDs []string `json:"location_ids"`
	Type        string   `json:"type"`
	// This field only need when upserting a notification from Backoffice
	Locations []InfoNotificationTarget_LocationFilter_Location `json:"locations"`
}

func (filter *InfoNotificationTarget_LocationFilter) GetNameValues() []string {
	names := []string{}
	for _, location := range filter.Locations {
		names = append(names, location.LocationName)
	}
	return names
}

// nolint
type InfoNotificationTarget_ClassFilter_Class struct {
	ClassID   string `json:"class_id"`
	ClassName string `json:"class_name"`
}

// nolint
type InfoNotificationTarget_ClassFilter struct {
	ClassIDs []string `json:"class_ids"`
	Type     string   `json:"type"`
	// This field only need when upserting a notification from Backoffice
	Classes []InfoNotificationTarget_ClassFilter_Class `json:"classes"`
}

func (filter *InfoNotificationTarget_ClassFilter) GetNameValues() []string {
	names := []string{}
	for _, class := range filter.Classes {
		names = append(names, class.ClassName)
	}
	return names
}

// nolint
type InfoNotificationTarget_SchoolFilter_School struct {
	SchoolID   string `json:"school_id"`
	SchoolName string `json:"school_name"`
}

// nolint
type InfoNotificationTarget_SchoolFilter struct {
	SchoolIDs []string `json:"school_ids"`
	Type      string   `json:"type"`
	// This field only need when upserting a notification from Backoffice
	Schools []InfoNotificationTarget_SchoolFilter_School `json:"schools"`
}

func (filter *InfoNotificationTarget_SchoolFilter) GetNameValues() []string {
	names := []string{}
	for _, school := range filter.Schools {
		names = append(names, school.SchoolName)
	}
	return names
}

// nolint
type InfoNotificationTarget_UserGroupFilter struct {
	UserGroups []string `json:"user_group"`
}

type InfoNotificationTarget struct {
	CourseFilter    InfoNotificationTarget_CourseFilter    `json:"course_filter"`
	GradeFilter     InfoNotificationTarget_GradeFilter     `json:"grade_filter"`
	UserGroupFilter InfoNotificationTarget_UserGroupFilter `json:"user_group_filter"`
	LocationFilter  InfoNotificationTarget_LocationFilter  `json:"location_filter"`
	ClassFilter     InfoNotificationTarget_ClassFilter     `json:"class_filter"`
	SchoolFilter    InfoNotificationTarget_SchoolFilter    `json:"school_filter"`
}

type GInfoNotificationFilter interface {
	GetNameValues() []string
}
