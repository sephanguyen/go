package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infra "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
)

type LessonMemberRepo struct {
	LessonmgmtLessonMemberRepo lesson_infra.LessonMemberRepo
}

func (l *LessonMemberRepo) GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, useUserBasicInfoTable bool) (map[string]lesson_domain.LessonLearners, error) {
	return l.LessonmgmtLessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx, db, lessonIDs, useUserBasicInfoTable)
}
