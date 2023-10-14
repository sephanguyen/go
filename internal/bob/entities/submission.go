package entities

import "github.com/jackc/pgtype"

// StudentSubmission reflects student_submissions table
type StudentSubmission struct {
	ID              pgtype.Text
	StudentID       pgtype.Text
	TopicID         pgtype.Text
	Content         pgtype.Text
	AttachmentNames pgtype.TextArray
	AttachmentURLs  pgtype.TextArray
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
}

// FieldMap returns StudentSubmission fields with names in SQL tables
func (s *StudentSubmission) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_submission_id",
			"student_id",
			"topic_id",
			"content",
			"attachment_names",
			"attachment_urls",
			"updated_at",
			"created_at",
		}, []interface{}{
			&s.ID,
			&s.StudentID,
			&s.TopicID,
			&s.Content,
			&s.AttachmentNames,
			&s.AttachmentURLs,
			&s.UpdatedAt,
			&s.CreatedAt,
		}
}

// TableName returns student_submissions
func (s *StudentSubmission) TableName() string {
	return "student_submissions"
}

// StudentSubmissionScore reflects student_submission_scores table
type StudentSubmissionScore struct {
	ID                  pgtype.Text
	TeacherID           pgtype.Text
	StudentSubmissionID pgtype.Text
	GivenScore          pgtype.Numeric
	TotalScore          pgtype.Numeric
	Notes               pgtype.Text
	CreatedAt           pgtype.Timestamptz
}

// FieldMap returns StudentSubmissionScore fields with names in SQL tables
func (s *StudentSubmissionScore) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_submission_score_id",
			"teacher_id",
			"student_submission_id",
			"given_score",
			"total_score",
			"notes",
			"created_at",
		}, []interface{}{
			&s.ID,
			&s.TeacherID,
			&s.StudentSubmissionID,
			&s.GivenScore,
			&s.TotalScore,
			&s.Notes,
			&s.CreatedAt,
		}
}

// TableName returns student_submission_scores
func (s *StudentSubmissionScore) TableName() string {
	return "student_submission_scores"
}
