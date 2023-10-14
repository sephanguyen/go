package entities

import (
	"github.com/jackc/pgtype"
)

// Student profiles and related information
type Student struct {
	User `sql:"-"`

	SchoolID          pgtype.Int4 `sql:"school_id"`
	School            *School     `sql:"-"`
	EnrollmentStatus  pgtype.Text `sql:"enrollment_status"`
	StudentExternalID pgtype.Text `sql:"student_external_id"`
	StudentNote       pgtype.Text `sql:"student_note"`

	ID                 pgtype.Text `sql:"student_id,pk"`
	CurrentGrade       pgtype.Int2
	TargetUniversity   pgtype.Text
	Biography          pgtype.Text
	Birthday           pgtype.Date
	TotalQuestionLimit pgtype.Int2        // Deprecated
	OnTrial            pgtype.Bool        `sql:",notnull"` // deprecated
	BillingDate        pgtype.Timestamptz // deprecated
	AdditionalData     pgtype.JSONB
	UpdatedAt          pgtype.Timestamptz
	CreatedAt          pgtype.Timestamptz
	DeletedAt          pgtype.Timestamptz
	GradeID            pgtype.Text
}

// FieldMap return a map of field name and pointer to field
func (e *Student) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id", "current_grade", "target_university", "biography", "birthday",
			"total_question_limit", "on_trial", "billing_date", "school_id", "enrollment_status",
			"student_external_id", "student_note", "additional_data", "updated_at", "created_at", "deleted_at", "grade_id",
		}, []interface{}{
			&e.ID, &e.CurrentGrade, &e.TargetUniversity, &e.Biography, &e.Birthday,
			&e.TotalQuestionLimit, &e.OnTrial, &e.BillingDate, &e.SchoolID, &e.EnrollmentStatus,
			&e.StudentExternalID, &e.StudentNote, &e.AdditionalData, &e.UpdatedAt, &e.CreatedAt, &e.DeletedAt, &e.GradeID,
		}
}

// TableName returns "students"
func (e *Student) TableName() string {
	return "students"
}

type StudentAdditionalData struct {
	JprefDivs []int64 `json:"jpref_divs,omitempty"`
}

func (e *Student) GetStudentAdditionalData() (*StudentAdditionalData, error) {
	r := &StudentAdditionalData{}
	err := e.AdditionalData.AssignTo(&r)
	return r, err
}

type StudentStat struct {
	StudentID           pgtype.Text
	TotalLOFinished     pgtype.Int4
	TotalLearningTime   pgtype.Int4
	LastTimeCompletedLO pgtype.Timestamptz // store last time a student completed a LO, used to calculate total learning time
	AdditionalData      pgtype.JSONB       // is_finished_first_lo, is_finished_first_topic boolean,
	UpdatedAt           pgtype.Timestamptz
	CreatedAt           pgtype.Timestamptz
}

func (s *StudentStat) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id", "total_lo_finished", "total_learning_time", "last_time_completed_lo", "additional_data", "updated_at", "created_at",
		}, []interface{}{
			&s.StudentID, &s.TotalLOFinished, &s.TotalLearningTime, &s.LastTimeCompletedLO, &s.AdditionalData, &s.UpdatedAt, &s.CreatedAt,
		}
}

func (s *StudentStat) TableName() string {
	return "student_statistics"
}
