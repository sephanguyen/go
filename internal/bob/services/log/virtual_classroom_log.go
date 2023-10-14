package log

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type VirtualClassRoomLogService struct {
	DB database.Ext

	Repo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.VirtualClassRoomLog) error
		AddAttendeeIDByLessonID(ctx context.Context, db database.QueryExecer, lessonID, attendeeID pgtype.Text) error
		IncreaseTotalTimesByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, logType entities.TotalTimes) error
		CompleteLogByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
		GetLatestByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*entities.VirtualClassRoomLog, error)
	}
}

func (v *VirtualClassRoomLogService) LogWhenAttendeeJoin(ctx context.Context, lessonID, attendeeID pgtype.Text) (createdNewLog bool, err error) {
	log, err := v.Repo.GetLatestByLessonID(ctx, v.DB, lessonID)
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
		e := &entities.VirtualClassRoomLog{}
		database.AllNullEntity(e)
		if err = multierr.Combine(
			e.LogID.Set(idutil.ULIDNow()),
			e.LessonID.Set(lessonID),
			e.IsCompleted.Set(false),
			e.AttendeeIDs.Set([]string{attendeeID.String}),
		); err != nil {
			return false, fmt.Errorf("could not set value for entity: %w", err)
		}

		if err = v.Repo.Create(ctx, v.DB, e); err != nil {
			return false, fmt.Errorf("Repo.Create: %w", err)
		}
	} else {
		if err = v.Repo.AddAttendeeIDByLessonID(ctx, v.DB, lessonID, attendeeID); err != nil {
			return false, fmt.Errorf("Repo.AddAttendeeIDByLessonID: %w", err)
		}
	}

	return createdNewLog, nil
}

func (v *VirtualClassRoomLogService) LogWhenUpdateRoomState(ctx context.Context, lessonID pgtype.Text) error {
	if err := v.Repo.IncreaseTotalTimesByLessonID(ctx, v.DB, lessonID, entities.TotalTimesUpdatingRoomState); err != nil {
		return fmt.Errorf("Repo.IncreaseTotalTimesByLessonID: %w", err)
	}

	return nil
}

func (v *VirtualClassRoomLogService) LogWhenGetRoomState(ctx context.Context, lessonID pgtype.Text) error {
	if err := v.Repo.IncreaseTotalTimesByLessonID(ctx, v.DB, lessonID, entities.TotalTimesGettingRoomState); err != nil {
		return fmt.Errorf("Repo.IncreaseTotalTimesByLessonID: %w", err)
	}

	return nil
}

func (v *VirtualClassRoomLogService) LogWhenEndRoom(ctx context.Context, lessonID pgtype.Text) error {
	if err := v.Repo.CompleteLogByLessonID(ctx, v.DB, lessonID); err != nil {
		return fmt.Errorf("Repo.CompleteLogByLessonID: %w", err)
	}

	return nil
}

func (v *VirtualClassRoomLogService) GetCompletedLogByLesson(ctx context.Context, lessonID pgtype.Text) (*entities.VirtualClassRoomLog, error) {
	log, err := v.Repo.GetLatestByLessonID(ctx, v.DB, lessonID)
	if err != nil {
		return nil, fmt.Errorf("Repo.GetLatestByLessonID: %w", err)
	}

	if !log.IsCompleted.Bool {
		return nil, nil
	}

	return log, nil
}
