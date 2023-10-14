package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgx/v4"
)

type DeleteLessonRequest struct {
	LessonID           string
	IsDeletedRecurring bool
}

func (l *LessonCommandHandler) DeleteLesson(ctx context.Context, req DeleteLessonRequest) ([]string, error) {
	var lessonIDs []string
	var err error
	var lesson *domain.Lesson
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if req.IsDeletedRecurring {
		lessonIDs, err = l.LessonRepo.GetFutureRecurringLessonIDs(ctx, conn, req.LessonID)
		if err != nil {
			return nil, err
		}
	} else {
		lessonIDs = append(lessonIDs, req.LessonID)
		lesson, err = l.LessonRepo.GetLessonWithSchedulerInfoByLessonID(ctx, conn, req.LessonID)
		if err != nil {
			return nil, fmt.Errorf("LessonRepo.GetLessonWithSchedulerInfoByLessonID: %w", err)
		}
	}
	if err := database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		if lesson != nil && lesson.SchedulerInfo != nil && lesson.SchedulerInfo.Freq == "once" && (lesson.ZoomID != "" || lesson.ClassDoRoomID != "") {
			if lesson.ZoomID != "" {
				_, err = l.ZoomService.RetryDeleteZoomLink(ctx, lesson.ZoomID)
				if err != nil {
					return fmt.Errorf("ZoomService.RetryDeleteZoomLink: %w", err)
				}
				err = l.LessonRepo.RemoveZoomLinkByLessonID(ctx, tx, req.LessonID)
				if err != nil {
					return fmt.Errorf("LessonRepo.RemoveZoomLinkByLessonID: %w", err)
				}
			} else {
				err = l.LessonRepo.RemoveClassDoLinkByLessonID(ctx, tx, req.LessonID)
				if err != nil {
					return fmt.Errorf("LessonRepo.RemoveClassDoLinkByLessonID: %w", err)
				}
			}
		}
		if err := l.LessonReportRepo.DeleteReportsBelongToLesson(ctx, tx, lessonIDs); err != nil {
			return fmt.Errorf("LessonReportRepo.DeleteReportsBelongToLesson: %w", err)
		}
		if err := l.LessonRepo.Delete(ctx, tx, lessonIDs); err != nil {
			return fmt.Errorf("LessonRepo.Delete: %w", err)
		}
		isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ReallocateStudents", l.Env)
		if err != nil {
			return fmt.Errorf("l.connectToUnleash: %w", err)
		}
		if isUnleashToggled {
			if err = l.ReallocationRepo.CancelReallocationByLessonID(ctx, tx, lessonIDs); err != nil {
				return fmt.Errorf("l.ReallocationRepo.CancelReallocationByLessonID: %w", err)
			}
			if err = l.ReallocationRepo.DeleteByOriginalLessonID(ctx, tx, lessonIDs); err != nil {
				return fmt.Errorf("l.ReallocationRepo.DeleteByOriginalLessonID: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return lessonIDs, nil
}
