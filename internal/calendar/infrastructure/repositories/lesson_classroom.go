package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infra "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
)

type LessonClassroomRepo struct {
	LessonmgmtLessonClassroom lesson_infra.LessonClassroomRepo
}

func (l *LessonClassroomRepo) GetLessonClassroomsWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]lesson_domain.LessonClassrooms, error) {
	return l.LessonmgmtLessonClassroom.GetLessonClassroomsWithNamesByLessonIDs(ctx, db, lessonIDs)
}
