package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_payloads "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infra "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
)

type LessonRepo struct {
	LessonmgmtLessonRepo lesson_infra.LessonRepo
}

func (l *LessonRepo) GetLessonWithNamesByID(ctx context.Context, db database.QueryExecer, lessonID string) (*lesson_domain.Lesson, error) {
	return l.LessonmgmtLessonRepo.GetLessonWithNamesByID(ctx, db, lessonID)
}

func (l *LessonRepo) GetLessonsByLocationStatusAndDateTimeRange(ctx context.Context, db database.QueryExecer, params *lesson_payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs) ([]*lesson_domain.Lesson, error) {
	return l.LessonmgmtLessonRepo.GetLessonsByLocationStatusAndDateTimeRange(ctx, db, params)
}
