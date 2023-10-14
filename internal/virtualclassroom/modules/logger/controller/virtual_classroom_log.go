package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type VirtualClassRoomLogService struct {
	WrapperConnection *support.WrapperDBConnection
	Repo              infrastructure.VirtualClassroomLogRepo
}

func (v *VirtualClassRoomLogService) LogWhenAttendeeJoin(ctx context.Context, lessonID, attendeeID string) (createdNewLog bool, err error) {
	conn, err := v.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	log, err := v.Repo.GetLatestByLessonID(ctx, conn, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			createdNewLog = true
		} else {
			return false, fmt.Errorf("Repo.GetLatestByLessonID: %w", err)
		}
	}

	// create new log
	if createdNewLog || log.IsCompleted.Bool {
		createdNewLog = true
		e := &repo.VirtualClassRoomLogDTO{}
		database.AllNullEntity(e)
		if err = multierr.Combine(
			e.LogID.Set(idutil.ULIDNow()),
			e.LessonID.Set(lessonID),
			e.IsCompleted.Set(false),
			e.AttendeeIDs.Set([]string{attendeeID}),
		); err != nil {
			return false, fmt.Errorf("could not set value for entity: %w", err)
		}

		if err = v.Repo.Create(ctx, conn, e); err != nil {
			return false, fmt.Errorf("Repo.Create: %w", err)
		}
	} else {
		if err = v.Repo.AddAttendeeIDByLessonID(ctx, conn, lessonID, attendeeID); err != nil {
			return false, fmt.Errorf("Repo.AddAttendeeIDByLessonID: %w", err)
		}
	}

	return createdNewLog, nil
}

func (v *VirtualClassRoomLogService) LogWhenUpdateRoomState(ctx context.Context, lessonID string) error {
	conn, err := v.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	if err := v.Repo.IncreaseTotalTimesByLessonID(ctx, conn, lessonID, repo.TotalTimesUpdatingRoomState); err != nil {
		return fmt.Errorf("Repo.IncreaseTotalTimesByLessonID: %w", err)
	}

	return nil
}

func (v *VirtualClassRoomLogService) LogWhenGetRoomState(ctx context.Context, lessonID string) error {
	conn, err :=  v.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	if err := v.Repo.IncreaseTotalTimesByLessonID(ctx, conn, lessonID, repo.TotalTimesGettingRoomState); err != nil {
		return fmt.Errorf("Repo.IncreaseTotalTimesByLessonID: %w", err)
	}

	return nil
}

func (v *VirtualClassRoomLogService) LogWhenEndRoom(ctx context.Context, lessonID string) error {
	conn, err :=  v.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	if err := v.Repo.CompleteLogByLessonID(ctx, conn, lessonID); err != nil {
		return fmt.Errorf("Repo.CompleteLogByLessonID: %w", err)
	}

	return nil
}

func (v *VirtualClassRoomLogService) GetCompletedLogByLesson(ctx context.Context, lessonID string) (*repo.VirtualClassRoomLogDTO, error) {
	conn, err :=  v.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	log, err := v.Repo.GetLatestByLessonID(ctx, conn, lessonID)
	if err != nil {
		return nil, fmt.Errorf("Repo.GetLatestByLessonID: %w", err)
	}

	if !log.IsCompleted.Bool {
		return nil, nil
	}

	return log, nil
}
