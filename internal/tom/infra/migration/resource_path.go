package migration

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	lentities "github.com/manabie-com/backend/internal/tom/domain/lesson"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

type ResourcePathMigrator struct {
	DB                     database.Ext
	ConversationLessonRepo interface {
		BulkUpdateResourcePath(ctx context.Context, db database.QueryExecer, lessons []string, resourcePath string) error
		FindByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray, includeSoftDeleted bool) ([]*lentities.ConversationLesson, error)
	}
	ConversationRepo interface {
		BulkUpdateResourcePath(ctx context.Context, db database.QueryExecer, convIDs []string, resourcePath string) error
	}
	UserDeviceTokenRepo interface {
		BulkUpdateResourcePath(ctx context.Context, db database.QueryExecer, userIDs []string, resourcePath string) error
	}
}

func (r *ResourcePathMigrator) MigrateUser(ctx context.Context, userInfo *tpb.ResourcePathMigration_Users) error {
	school := userInfo.GetSchoolId()

	err := r.UserDeviceTokenRepo.BulkUpdateResourcePath(ctx, r.DB, userInfo.GetUserIds(), school)
	if err != nil {
		return fmt.Errorf("r.UserDeviceTokenRepo.BulkUpdateUserToken: %w", err)
	}
	return nil
}

func (r *ResourcePathMigrator) MigrateLesson(ctx context.Context, lessonInfo *tpb.ResourcePathMigration_Lessons) error {
	school := lessonInfo.GetSchoolId()
	lessonIDs := lessonInfo.GetLessonIds()
	if len(lessonIDs) == 0 {
		return nil
	}

	err := database.ExecInTx(ctx, r.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := r.ConversationLessonRepo.BulkUpdateResourcePath(ctx, tx, lessonIDs, school)
		if err != nil {
			return fmt.Errorf("r.UserDeviceTokenRepo.BulkUpdateUserToken: %w", err)
		}
		convLessons, err := r.ConversationLessonRepo.FindByLessonIDs(ctx, tx, database.TextArray(lessonIDs), true)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
		}
		convIDs := make([]string, 0, len(convLessons))
		for _, convLesson := range convLessons {
			convIDs = append(convIDs, convLesson.ConversationID.String)
		}
		err = r.ConversationRepo.BulkUpdateResourcePath(ctx, tx, convIDs, school)
		if err != nil {
			return fmt.Errorf("r.ConversationRepo.BulkUpdateResourcePath %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}
	return nil
}
