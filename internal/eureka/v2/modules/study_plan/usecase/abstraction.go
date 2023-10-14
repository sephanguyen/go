package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type StudyPlanItemUseCase struct {
	DB                   database.Ext
	LearningMaterialRepo repository.LearningMaterialRepo
	StudyPlanItemRepo    repository.StudyPlanItemRepo
}

func NewStudyPlanItemUseCase(db database.Ext, studyPlanItemRepo repository.StudyPlanItemRepo, learningMaterialListRepo repository.LearningMaterialRepo) *StudyPlanItemUseCase {
	return &StudyPlanItemUseCase{
		DB:                   db,
		StudyPlanItemRepo:    studyPlanItemRepo,
		LearningMaterialRepo: learningMaterialListRepo,
	}
}

type UpsertStudyPlanItem interface {
	UpsertStudyPlanItems(ctx context.Context, courses []domain.StudyPlanItem) error
}

type StudyPlanUsecase interface {
	UpsertStudyPlan(ctx context.Context, studyPlan domain.StudyPlan) (string, error)
}

type StudyPlanUsecaseImpl struct {
	DB database.Ext

	StudyPlanRepo repository.StudyPlanRepo
}

func NewStudyPlanUsecase(db database.Ext, studyPlanRepo *postgres.StudyPlanRepo) StudyPlanUsecase {
	return &StudyPlanUsecaseImpl{
		DB:            db,
		StudyPlanRepo: studyPlanRepo,
	}
}
