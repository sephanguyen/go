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

const LESSON_MEMBERS_EXECUTOR = LabelExecutor("LESSON_MEMBERS_EXECUTOR")

type LessonMemberExecutor struct {
	service          *application.UpdaterLessonMember
	db               database.Ext
	logger           *zap.Logger
	LessonMemberRepo infrastructure.LessonMemberRepo
}

func InitLessonMembersExecutor(db database.Ext, logger *zap.Logger) *LessonMemberExecutor {
	lessonMemberRepo := &lesson_repo.LessonMemberRepo{}
	userRepo := &lesson_user_repo.UserRepo{}
	return &LessonMemberExecutor{
		db:               db,
		logger:           logger,
		LessonMemberRepo: lessonMemberRepo,
		service: &application.UpdaterLessonMember{
			DB:               db,
			LessonMemberRepo: lessonMemberRepo,
			UserRepo:         userRepo,
		},
	}
}

func (l *LessonMemberExecutor) GetTotal(ctx context.Context) (int, error) {
	query := "SELECT COUNT(1) FROM lesson_members WHERE resource_path = $1"
	var totalLesson int
	if err := l.db.QueryRow(ctx, query, golibs.ResourcePathFromCtx(ctx)).Scan(&totalLesson); err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("row.Scan: %w", err)
	}
	return totalLesson, nil
}

func (l *LessonMemberExecutor) ExecuteJob(ctx context.Context, limit int, offSet int) error {
	resourceID := golibs.ResourcePathFromCtx(ctx)

	lessonMembers, err := l.LessonMemberRepo.FindByResourcePath(ctx, l.db, resourceID, limit, offSet)
	if err != nil {
		return fmt.Errorf("fail Sync Lesson Member in org: %s, offset: %d, limit: %d: %s", resourceID, offSet, limit, err)
	}
	err = l.service.UpdateLessonMemberNames(ctx, lessonMembers)
	if err != nil {
		return fmt.Errorf("fail update lesson member: %s", err)
	}

	l.logger.Info(fmt.Sprintf("the total of sync lessons members success in org: %s, offset: %d, limit: %d: %d", resourceID, offSet, limit, len(*lessonMembers)))
	return nil
}

func (l *LessonMemberExecutor) PreExecute(ctx context.Context) error {
	// do something
	return nil
}
