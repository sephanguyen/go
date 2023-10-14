package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infra "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
)

type LessonTeacherRepo struct {
	LessonmgmtLessonTeacherRepo lesson_infra.LessonTeacherRepo
}

func (l *LessonTeacherRepo) GetTeachersWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, useUserBasicInfoTable bool) (map[string]lesson_domain.LessonTeachers, error) {
	return l.LessonmgmtLessonTeacherRepo.GetTeachersWithNamesByLessonIDs(ctx, db, lessonIDs, useUserBasicInfoTable)
}

func (l *LessonTeacherRepo) GetTeachersByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]lesson_domain.LessonTeachers, error) {
	return l.LessonmgmtLessonTeacherRepo.GetTeachersByLessonIDs(ctx, db, lessonIDs)
}
