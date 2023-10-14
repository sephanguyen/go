package courses

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type PresetStudyPlanBuilder struct {
	db         database.Ext
	repo       repo.PresetStudyPlanRepo
	courseRepo repo.CourseRepo
}

func NewPresetStudyPlanBuilder(db database.Ext, repo repo.PresetStudyPlanRepo, courseRepo repo.CourseRepo) *PresetStudyPlanBuilder {
	return &PresetStudyPlanBuilder{
		db:         db,
		repo:       repo,
		courseRepo: courseRepo,
	}
}

// CreatePresetStudyPlansByCourseIDs creates a preset_study_plan for each course that
// does not have a preset_study_plan yet. A preset_study_plan contains data such as:
// country, grade, subject and start date for course.
func (p *PresetStudyPlanBuilder) CreatePresetStudyPlansByCourseIDs(ctx context.Context, courseIDs pgtype.TextArray) error {
	courses, err := p.courseRepo.FindByIDs(ctx, p.db, courseIDs)
	if err != nil {
		return err
	}

	coursesNotHavePSP := make([]*entities.Course, 0, len(courses))
	for i := range courses {
		if courses[i].PresetStudyPlanID.Status == pgtype.Present && len(courses[i].PresetStudyPlanID.String) != 0 {
			continue
		}
		coursesNotHavePSP = append(coursesNotHavePSP, courses[i])
	}
	if len(coursesNotHavePSP) == 0 {
		return nil
	}

	presetStudyPlans := make([]*entities.PresetStudyPlan, 0, len(coursesNotHavePSP))
	for i, course := range coursesNotHavePSP {
		presetStudyPlan := &entities.PresetStudyPlan{}
		database.AllNullEntity(presetStudyPlan)
		err = multierr.Combine(
			presetStudyPlan.ID.Set(idutil.ULIDNow()),
			presetStudyPlan.Country.Set(course.Country.String),
			presetStudyPlan.Name.Set(course.Name.String),
			presetStudyPlan.Grade.Set(course.Grade.Int),
			presetStudyPlan.Subject.Set(course.Subject.String),
			presetStudyPlan.StartDate.Set(course.StartDate.Time),
		)
		if err != nil {
			return fmt.Errorf("create preset study plan by course: %w", err)
		}

		presetStudyPlans = append(presetStudyPlans, presetStudyPlan)
		coursesNotHavePSP[i].PresetStudyPlanID = presetStudyPlan.ID
	}

	err = p.repo.CreatePresetStudyPlan(ctx, p.db, presetStudyPlans)
	if err != nil {
		return fmt.Errorf("PresetStudyPlanRepo.CreatePresetStudyPlan: %w", err)
	}

	// update preset study plan id field of courses
	err = p.courseRepo.Upsert(ctx, p.db, coursesNotHavePSP)
	if err != nil {
		return fmt.Errorf("CourseRepo.Upsert: %w", err)
	}
	return nil
}
