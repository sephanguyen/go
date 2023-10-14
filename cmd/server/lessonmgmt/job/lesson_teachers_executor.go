package job

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lesson_user_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

const LESSON_TEACHERS_EXECUTOR = LabelExecutor("LESSON_TEACHERS_EXECUTOR")

type LessonTeacherExecutor struct {
	service           *application.UpdaterLessonTeacher
	db                database.Ext
	logger            *zap.Logger
	LessonTeacherRepo infrastructure.LessonTeacherRepo
	UserRepo          *lesson_user_repo.UserRepo
}

func InitLessonTeacherExecutor(db database.Ext, logger *zap.Logger) *LessonTeacherExecutor {
	lessonTeacherRepo := &lesson_repo.LessonTeacherRepo{}
	userRepo := &lesson_user_repo.UserRepo{}
	return &LessonTeacherExecutor{
		db:                db,
		logger:            logger,
		UserRepo:          userRepo,
		LessonTeacherRepo: lessonTeacherRepo,
		service: &application.UpdaterLessonTeacher{
			DB:                db,
			LessonTeacherRepo: lessonTeacherRepo,
		},
	}
}

func (l *LessonTeacherExecutor) GetTotal(ctx context.Context) (int, error) {
	query := "select COUNT(1) from users WHERE resource_path = $1 AND deleted_at is null"

	var totalLesson int
	if err := l.db.QueryRow(ctx, query, golibs.ResourcePathFromCtx(ctx)).Scan(&totalLesson); err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("row.Scan: %w", err)
	}
	return totalLesson, nil
}

func (l *LessonTeacherExecutor) ExecuteJob(ctx context.Context, limit int, offSet int) error {
	resourceID := golibs.ResourcePathFromCtx(ctx)

	teachers, err := l.UserRepo.FindByResourcePath(ctx, l.db, resourceID, limit, offSet)
	if err != nil {
		return fmt.Errorf("fail Sync Lesson Teacher in org: %s, offset: %d, limit: %d: %s", resourceID, offSet, limit, err)
	}
	err = l.service.UpdaterLessonTeacherNames(ctx, teachers)
	if err != nil {
		return fmt.Errorf("fail update lesson teacher: %s", err)
	}

	l.logger.Info(fmt.Sprintf("the total of sync lessons teachers success in org: %s, offset: %d, limit: %d: %d", resourceID, offSet, limit, len(teachers)))
	return nil
}

func (l *LessonTeacherExecutor) PreExecute(ctx context.Context) error {
	// do something
	return nil
}
