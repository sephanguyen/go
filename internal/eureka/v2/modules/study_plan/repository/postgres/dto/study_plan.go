package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type StudyPlan struct {
	BaseEntity
	ID           pgtype.Text
	Name         pgtype.Text
	CourseID     pgtype.Text
	AcademicYear pgtype.Text
	Status       pgtype.Text
}

func (a *StudyPlan) FieldMap() ([]string, []interface{}) {
	return []string{
			"study_plan_id",
			"name",
			"course_id",
			"academic_year",
			"status",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&a.ID,
			&a.Name,
			&a.CourseID,
			&a.AcademicYear,
			&a.Status,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		}
}

func (a *StudyPlan) TableName() string {
	return "lms_study_plans"
}

func (a *StudyPlan) ToEntity() (domain.StudyPlan, error) {
	studyPlan, err := domain.NewStudyPlan(domain.StudyPlan{
		ID:           a.ID.String,
		Name:         a.Name.String,
		AcademicYear: a.AcademicYear.String,
		CourseID:     a.CourseID.String,
		Status:       domain.StudyPlanStatus(a.Status.String),
	})
	if err != nil {
		return domain.StudyPlan{}, errors.NewConversionError("domain.NewStudyPlan", err)
	}

	return studyPlan, nil
}

func (a *StudyPlan) FromEntity(now time.Time, d domain.StudyPlan) error {
	database.AllNullEntity(a)

	if err := multierr.Combine(
		a.ID.Set(d.ID),
		a.CourseID.Set(d.CourseID),
		a.Name.Set(d.Name),
		a.AcademicYear.Set(d.AcademicYear),
		a.Status.Set(d.Status),
		a.CreatedAt.Set(now),
		a.UpdatedAt.Set(now),
	); err != nil {
		return errors.NewConversionError("multierr.Combine", err)
	}

	return nil
}
