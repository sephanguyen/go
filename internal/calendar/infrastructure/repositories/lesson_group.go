package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infra "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
)

type LessonGroupRepo struct {
	LessonmgmtLessonGroupRepo lesson_infra.LessonGroupRepo
}

func (l *LessonGroupRepo) ListMediaByLessonArgs(ctx context.Context, db database.QueryExecer, args *lesson_domain.ListMediaByLessonArgs) (media_domain.Medias, error) {
	return l.LessonmgmtLessonGroupRepo.ListMediaByLessonArgs(ctx, db, args)
}
