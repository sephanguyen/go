package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type StudyPlan struct {
	BaseEntity
	ID                  pgtype.Text
	MasterStudyPlan     pgtype.Text
	Name                pgtype.Text
	StudyPlanType       pgtype.Text
	SchoolID            pgtype.Int4
	CourseID            pgtype.Text
	BookID              pgtype.Text
	Status              pgtype.Text
	TrackSchoolProgress pgtype.Bool
	Grades              pgtype.Int4Array
}

func (rcv *StudyPlan) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"study_plan_id",
		"master_study_plan_id",
		"name",
		"study_plan_type",
		"created_at",
		"updated_at",
		"deleted_at",
		"school_id",
		"course_id",
		"book_id",
		"status",
		"track_school_progress",
		"grades",
	}
	values = []interface{}{
		&rcv.ID,
		&rcv.MasterStudyPlan,
		&rcv.Name,
		&rcv.StudyPlanType,
		&rcv.CreatedAt,
		&rcv.UpdatedAt,
		&rcv.DeletedAt,
		&rcv.SchoolID,
		&rcv.CourseID,
		&rcv.BookID,
		&rcv.Status,
		&rcv.TrackSchoolProgress,
		&rcv.Grades,
	}
	return
}

func (rcv *StudyPlan) TableName() string {
	return "study_plans"
}

type StudyPlans []*StudyPlan

func (ss *StudyPlans) Add() database.Entity {
	e := &StudyPlan{}
	*ss = append(*ss, e)

	return e
}

type StudyPlanCombineStudentID struct {
	StudentID pgtype.Text
	StudyPlan
}
