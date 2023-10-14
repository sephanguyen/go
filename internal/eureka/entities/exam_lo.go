package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ExamLO struct {
	LearningMaterial
	Instruction    pgtype.Text
	GradeToPass    pgtype.Int4
	ManualGrading  pgtype.Bool
	TimeLimit      pgtype.Int4
	ApproveGrading pgtype.Bool
	GradeCapping   pgtype.Bool
	MaximumAttempt pgtype.Int4
	ReviewOption   pgtype.Text
}

type ExamLOBase struct {
	ExamLO
	TotalQuestion pgtype.Int4
}

type GradeBookSetting struct {
	BaseEntity
	Setting   pgtype.Text
	UpdatedBy pgtype.Text
}

type ExamLoScore struct {
	CourseID              pgtype.Text
	StudyPlanID           pgtype.Text
	StudyPlanName         pgtype.Text
	StudentID             pgtype.Text
	StudentName           pgtype.Text
	LearningMaterialID    pgtype.Text
	ExamLOName            pgtype.Text
	Status                pgtype.Text
	IsGradeToPass         pgtype.Bool
	Grade                 pgtype.Int2
	GradeID               pgtype.Text
	GradePoint            pgtype.Int2
	TotalPoint            pgtype.Int2
	PassedExamLo          pgtype.Bool
	TotalAttempts         pgtype.Int2
	TotalExamLOs          pgtype.Int8
	TotalCompletedExamLOs pgtype.Int8
	TotalGradeToPass      pgtype.Int8
	TotalPassed           pgtype.Int8
	ReviewOption          pgtype.Text
	ChapterDisplayOrder   pgtype.Int2
	TopicDisplayOrder     pgtype.Int2
	LmDisplayOrder        pgtype.Int2
	DueDate               pgtype.Timestamptz
}

type StudentStudyPlanIdentity struct {
	StudentID     string
	StudentName   string
	Grade         int32
	CourseID      string
	StudyPlanID   string
	StudyPlanName string
}

func (t *GradeBookSetting) FieldMap() ([]string, []interface{}) {
	return []string{
			"setting",
			"updated_by",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&t.Setting,
			&t.UpdatedBy,
			&t.UpdatedAt,
			&t.CreatedAt,
			&t.DeletedAt,
		}
}

func (t *GradeBookSetting) TableName() string {
	return "grade_book_setting"
}

func (t *ExamLO) FieldMap() ([]string, []interface{}) {
	fields, values := t.LearningMaterial.FieldMap()

	fields = append(fields, "instruction", "grade_to_pass", "manual_grading", "time_limit", "maximum_attempt", "approve_grading", "grade_capping", "review_option")
	values = append(values, &t.Instruction, &t.GradeToPass, &t.ManualGrading, &t.TimeLimit, &t.MaximumAttempt, &t.ApproveGrading, &t.GradeCapping, &t.ReviewOption)

	return fields, values
}

func (t *ExamLO) TableName() string {
	return "exam_lo"
}

type ExamLOs []*ExamLO

func (u *ExamLOs) Add() database.Entity {
	e := &ExamLO{}
	*u = append(*u, e)

	return e
}

func (u ExamLOs) Get() []*ExamLO {
	return []*ExamLO(u)
}
