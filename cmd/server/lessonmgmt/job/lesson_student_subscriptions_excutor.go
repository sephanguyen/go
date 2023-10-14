package job

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	lesson_user_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

const LESSON_STUDENT_SUBSCRIPTIONS_EXECUTOR = LabelExecutor("LESSON_STUDENT_SUBSCRIPTIONS_EXECUTOR")

type LessonStudentSubscriptionExecutor struct {
	service                       *application.UpdaterLessonStudentSubscription
	db                            database.Ext
	logger                        *zap.Logger
	LessonStudentSubscriptionRepo *lesson_user_repo.StudentSubscriptionRepo
	UserRepo                      *lesson_user_repo.UserRepo
}

func InitLessonStudentSubscriptionExecutor(db database.Ext, logger *zap.Logger) *LessonStudentSubscriptionExecutor {
	lessonStudentSubscriptionRepo := &lesson_user_repo.StudentSubscriptionRepo{}
	userRepo := &lesson_user_repo.UserRepo{}
	return &LessonStudentSubscriptionExecutor{
		db:       db,
		logger:   logger,
		UserRepo: userRepo,
		service: &application.UpdaterLessonStudentSubscription{
			DB:                      db,
			StudentSubscriptionRepo: lessonStudentSubscriptionRepo,
		},
	}
}

func (l *LessonStudentSubscriptionExecutor) GetTotal(ctx context.Context) (int, error) {
	query := "select COUNT(1) from users WHERE resource_path = $1 AND deleted_at is null"

	var totalLesson int
	if err := l.db.QueryRow(ctx, query, golibs.ResourcePathFromCtx(ctx)).Scan(&totalLesson); err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("row.Scan: %w", err)
	}
	return totalLesson, nil
}

func (l *LessonStudentSubscriptionExecutor) ExecuteJob(ctx context.Context, limit int, offSet int) error {
	resourceID := golibs.ResourcePathFromCtx(ctx)

	students, err := l.UserRepo.FindByResourcePath(ctx, l.db, resourceID, limit, offSet)
	if err != nil {
		return fmt.Errorf("fail Sync Student Subscription in org: %s, offset: %d, limit: %d: %s", resourceID, offSet, limit, err)
	}
	err = l.service.UpdateStudentNamesOfStudentSubscription(ctx, students)
	if err != nil {
		return fmt.Errorf("fail update student subscription: %s", err)
	}

	l.logger.Info(fmt.Sprintf("the total of sync student subscription success in org: %s, offset: %d, limit: %d: %d", resourceID, offSet, limit, len(students)))
	return nil
}

func (l *LessonStudentSubscriptionExecutor) PreExecute(ctx context.Context) error {
	// do something
	return nil
}
