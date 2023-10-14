package entities

import (
	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StudentLearningTimeDaily struct {
	ID                      pgtype.Int4
	StudentID               pgtype.Text
	LearningTime            pgtype.Int4
	Day                     pgtype.Timestamptz
	Sessions                pgtype.Text
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	AssignmentLearningTime  pgtype.Int4
	AssignmentSubmissionIDs pgtype.TextArray
}

type StudentAssignmentLearningTime struct {
	StudentID    string
	AssignmentID string
	CompleteDate *timestamppb.Timestamp
	Duration     int32
}

func (s *StudentLearningTimeDaily) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"learning_time_id",
		"student_id",
		"learning_time",
		"day",
		"sessions",
		"created_at",
		"updated_at",
		"assignment_learning_time",
		"assignment_submission_ids",
	}

	values = []interface{}{
		&s.ID,
		&s.StudentID,
		&s.LearningTime,
		&s.Day,
		&s.Sessions,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.AssignmentLearningTime,
		&s.AssignmentSubmissionIDs,
	}
	return
}

func (s *StudentLearningTimeDaily) TableName() string {
	return "student_learning_time_by_daily"
}
