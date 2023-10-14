package entity

import (
	"time"

	"github.com/jackc/pgtype"
)

// struct type
type HasuraAssignmentContent struct {
	pgtype.JSONB
	TopicID string   `graphql:"topic_id"`
	LoID    []string `graphql:"lo_id"`
}
type HasuraAssignmentCheckList struct {
	pgtype.JSONB
	Items []struct {
		Content   string `graphql:"content"`
		IsChecked bool   `graphql:"is_checked"`
	} `graphql:"items"`
}
type HasuraAssignmentSetting struct {
	pgtype.JSONB
	AllowLateSubmission bool `graphql:"allow_late_submission"`
	AllowResubmission   bool `graphql:"allow_resubmission"`
	RequireAttachment   bool `graphql:"require_attachment"`
}
type AssignmentAttrs struct {
	AssignmentID string `graphql:"assignment_id"`
	Instruction  string `graphql:"instruction"`

	Content   HasuraAssignmentContent   `graphql:"content"`
	Settings  HasuraAssignmentSetting   `graphql:"settings"`
	CheckList HasuraAssignmentCheckList `graphql:"check_list"`

	Attachment      []string  `graphql:"attachment"`
	Types           string    `graphql:"type"`
	Name            string    `graphql:"name"`
	MaxGrade        int32     `graphql:"max_grade"`
	IsRequiredGrade bool      `graphql:"is_required_grade"`
	CreatedAt       time.Time `graphql:"created_at"`
	DisplayOrder    int32     `graphql:"display_order"`
}

// Query
type GraphqlAssignmentsByTopicIDQuery struct {
	Assignments []struct {
		Name         string                  `graphql:"name"`
		AssignmentID string                  `graphql:"assignment_id"`
		Content      HasuraAssignmentContent `graphql:"content"`
		CreatedAt    time.Time               `graphql:"created_at"`
		DisplayOrder int32                   `graphql:"display_order"`
	} `graphql:"find_assignment_by_topic_id(args: {ids: $topic_id}, order_by: [$order_by])"`
}

type GraphqlAssignmentsManyQuery struct {
	Assignments []struct {
		AssignmentAttrs `graphql:"...on assignments"`
	} `graphql:"assignments(order_by: {display_order: asc}, where: {assignment_id: {_in: $assignment_id}})"`
}

type GraphqlAssignmentOneQuery struct {
	Assignments []struct {
		AssignmentAttrs `graphql:"...on assignments"`
	} `graphql:"assignments(where: {assignment_id: {_eq: $assignment_id}})"`
}

type GraphqlAssignmentDisplayOrder struct {
	Assignments []struct {
		DisplayOrder int32 `graphql:"display_order"`
	} `graphql:"find_assignment_by_topic_id( args: { ids: $topic_id } order_by: { display_order: desc }	limit: 1)"`
}
