package usecase

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

func (a *StudyPlanUsecaseImpl) UpsertStudyPlan(ctx context.Context, studyPlan domain.StudyPlan) (string, error) {
	id, err := a.StudyPlanRepo.Upsert(ctx, a.DB, time.Now(), domain.StudyPlan{
		ID:           idutil.ULIDNow(),
		Name:         studyPlan.Name,
		CourseID:     studyPlan.CourseID,
		AcademicYear: studyPlan.AcademicYear,
		Status:       studyPlan.Status,
	})
	if err != nil {
		return "", errors.New("StudyPlanUsecase.StudyPlanRepo.Upsert", err)
	}

	return id, nil
}
