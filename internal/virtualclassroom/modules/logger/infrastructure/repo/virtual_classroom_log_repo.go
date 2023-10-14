package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/pkg/errors"
)

type VirtualClassroomLogRepo struct{}

const getLatestLogIDByLessonIDQuery = `SELECT log_id FROM virtual_classroom_log 
WHERE lesson_id = $1 and deleted_at is null and is_completed = FALSE 
ORDER BY created_at DESC LIMIT 1`

func (v *VirtualClassroomLogRepo) Create(ctx context.Context, db database.QueryExecer, e *VirtualClassRoomLogDTO) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualClassroomLogRepo.Create")
	defer span.End()

	now := time.Now()
	_ = e.UpdatedAt.Set(now)
	_ = e.CreatedAt.Set(now)
	_ = e.DeletedAt.Set(nil)

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new VirtualClassroomLog")
	}

	return nil
}

func (v *VirtualClassroomLogRepo) AddAttendeeIDByLessonID(ctx context.Context, db database.QueryExecer, lessonID, attendeeID string) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualClassroomLogRepo.AddAttendeeIDByLessonID")
	defer span.End()

	query := fmt.Sprintf(`UPDATE virtual_classroom_log SET attendee_ids = array_append(attendee_ids, $2), updated_at = now() 
WHERE NOT($2 = ANY(attendee_ids)) AND log_id IN (%s)`, getLatestLogIDByLessonIDQuery)
	_, err := db.Exec(ctx, query, lessonID, attendeeID)
	if err != nil {
		return err
	}

	return nil
}

func (v *VirtualClassroomLogRepo) IncreaseTotalTimesByLessonID(ctx context.Context, db database.QueryExecer, lessonID string, logType TotalTimes) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualClassroomLogRepo.IncreaseTotalTimesByLessonID")
	defer span.End()

	var willUpdateField string
	switch logType {
	case TotalTimesReconnection:
		willUpdateField = "total_times_reconnection"
	case TotalTimesUpdatingRoomState:
		willUpdateField = "total_times_updating_room_state"
	case TotalTimesGettingRoomState:
		willUpdateField = "total_times_getting_room_state"
	default:
		return fmt.Errorf("not handle this type yet %v", logType)
	}

	query := fmt.Sprintf(`UPDATE virtual_classroom_log SET :willUpdateField = coalesce(:willUpdateField, 0) + 1, updated_at = now() 
WHERE log_id IN (%s)`, getLatestLogIDByLessonIDQuery)
	query = strings.ReplaceAll(query, ":willUpdateField", willUpdateField)
	_, err := db.Exec(ctx, query, lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (v *VirtualClassroomLogRepo) CompleteLogByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) error {
	ctx, span := interceptors.StartSpan(ctx, "VirtualClassroomLogRepo.CompleteLogByLessonID")
	defer span.End()

	query := fmt.Sprintf(`UPDATE virtual_classroom_log SET is_completed = TRUE, updated_at = now() WHERE log_id IN (%s)`, getLatestLogIDByLessonIDQuery)
	_, err := db.Exec(ctx, query, lessonID)
	if err != nil {
		return err
	}

	return nil
}

func (v *VirtualClassroomLogRepo) GetLatestByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (*VirtualClassRoomLogDTO, error) {
	ctx, span := interceptors.StartSpan(ctx, "VirtualClassroomLogRepo.GetLatestByLessonID")
	defer span.End()

	e := &VirtualClassRoomLogDTO{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = $1 and deleted_at is null ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &lessonID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}
