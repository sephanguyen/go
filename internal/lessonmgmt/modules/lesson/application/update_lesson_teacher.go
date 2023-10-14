package application

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
)

type UpdaterLessonTeacher struct {
	DB database.Ext

	// ports
	LessonTeacherRepo infrastructure.LessonTeacherRepo
}

func (l *UpdaterLessonTeacher) UpdaterLessonTeacherNames(ctx context.Context, users user_domain.Users) error {
	updateLessonTeachers := make([]*domain.UpdateLessonTeacherName, 0, len(users))

	for _, user := range users {
		updateLessonTeacher := domain.UpdateLessonTeacherName{
			TeacherID: user.ID,
			FullName:  user.FullName,
		}
		updateLessonTeachers = append(updateLessonTeachers, &updateLessonTeacher)
	}

	if err := l.LessonTeacherRepo.UpdateLessonTeacherNames(ctx, l.DB, updateLessonTeachers); err != nil {
		return fmt.Errorf("UpdaterLessonTeacher.UpdateLessonTeacherNames err: %w", err)
	}

	return nil
}
