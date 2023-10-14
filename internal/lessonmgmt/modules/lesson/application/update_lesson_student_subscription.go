package application

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
)

type UpdaterLessonStudentSubscription struct {
	DB database.Ext

	// ports
	StudentSubscriptionRepo infrastructure.StudentSubscriptionRepo
}

func (l *UpdaterLessonStudentSubscription) UpdateStudentNamesOfStudentSubscription(ctx context.Context, users user_domain.Users) error {
	if err := l.StudentSubscriptionRepo.UpdateMultiStudentNameByStudents(ctx, l.DB, users); err != nil {
		return fmt.Errorf("UpdaterLessonStudentSubscription.UpdaterLessonStudentNamesOfStudentSubscription err: %w", err)
	}

	return nil
}
