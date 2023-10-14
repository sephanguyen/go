package repository

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type StudyPlanItemRepo interface {
	UpsertStudyPlanItems(ctx context.Context, studyPlanItems []*dto.StudyPlanItemDto) error
}

type LearningMaterialRepo interface {
	UpsertLearningMaterialsIDList(ctx context.Context, lmLists []*dto.LmListDto) error
}

type StudyPlanRepo interface {
	Upsert(ctx context.Context, db database.Ext, now time.Time, studyPlan domain.StudyPlan) (string, error)
}
