package entities

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type PresetStudyPlan struct {
	ID        pgtype.Text `sql:"preset_study_plan_id,pk"`
	Name      pgtype.Text
	Country   pgtype.Text
	Grade     pgtype.Int2
	Subject   pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	StartDate pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (p *PresetStudyPlan) FieldMap() ([]string, []interface{}) {
	return []string{
			"preset_study_plan_id", "name", "country", "grade", "subject", "updated_at", "created_at", "start_date", "deleted_at",
		}, []interface{}{
			&p.ID, &p.Name, &p.Country, &p.Grade, &p.Subject, &p.UpdatedAt, &p.CreatedAt, &p.StartDate, &p.DeletedAt,
		}
}

func (p *PresetStudyPlan) TableName() string {
	return "preset_study_plans"
}

type PresetStudyPlanWeekly struct {
	ID                pgtype.Text `sql:"preset_study_plan_weekly_id,pk"`
	PresetStudyPlanID pgtype.Text `sql:"preset_study_plan_id"`
	TopicID           pgtype.Text `sql:"topic_id"`
	Week              pgtype.Int2 `sql:"week"`
	LessonID          pgtype.Text `sql:"lesson_id"`
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	StartDate         pgtype.Timestamptz
	EndDate           pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (p *PresetStudyPlanWeekly) FieldMap() ([]string, []interface{}) {
	return []string{
			"preset_study_plan_weekly_id", "preset_study_plan_id", "topic_id", "week", "lesson_id", "updated_at", "created_at", "start_date", "end_date", "deleted_at",
		}, []interface{}{
			&p.ID, &p.PresetStudyPlanID, &p.TopicID, &p.Week, &p.LessonID, &p.UpdatedAt, &p.CreatedAt, &p.StartDate, &p.EndDate, &p.DeletedAt,
		}
}

func (p *PresetStudyPlanWeekly) TableName() string {
	return "preset_study_plans_weekly"
}

func (p *PresetStudyPlanWeekly) IsValid() error {
	if p.PresetStudyPlanID.Status != pgtype.Present {
		return fmt.Errorf("preset study plan id cannot be empty")
	}

	if p.TopicID.Status != pgtype.Present {
		return fmt.Errorf("topic id cannot be empty")
	}

	if p.Week.Status != pgtype.Present {
		return fmt.Errorf("week cannot be empty")
	}

	if p.StartDate.Status != pgtype.Present || p.StartDate.Time.IsZero() {
		return fmt.Errorf("start date cannot be empty")
	}

	if p.EndDate.Status != pgtype.Present || p.EndDate.Time.IsZero() {
		return fmt.Errorf("end date cannot be empty")
	}

	if p.EndDate.Time.Before(p.StartDate.Time) {
		return fmt.Errorf("start date must before end date")
	}

	return nil
}

type PresetStudyPlansWeekly []*PresetStudyPlanWeekly

func (u *PresetStudyPlansWeekly) Add() database.Entity {
	e := &PresetStudyPlanWeekly{}
	*u = append(*u, e)

	return e
}

type StudentsStudyPlansWeekly struct {
	PresetStudyPlanWeeklyID pgtype.Text
	StudentID               pgtype.Text
	StartDate               pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
}

func (p *StudentsStudyPlansWeekly) FieldMap() ([]string, []interface{}) {
	return []string{
			"preset_study_plan_weekly_id", "student_id", "start_date", "updated_at", "created_at",
		}, []interface{}{
			&p.PresetStudyPlanWeeklyID, &p.StudentID, &p.StartDate, &p.UpdatedAt, &p.CreatedAt,
		}
}

func (*StudentsStudyPlansWeekly) TableName() string {
	return "students_study_plans_weekly"
}
