package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type LessonClassroomRepo struct{}

func (l *LessonClassroomRepo) GetClassroomIDsByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*LessonClassroom, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonClassroomRepo.GetClassroomIDsByLessonIDs")
	defer span.End()

	lessonClassroom := &LessonClassroom{}
	fields, _ := lessonClassroom.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE lesson_id = ANY($1) AND deleted_at is null`,
		strings.Join(fields, ","),
		lessonClassroom.TableName())
	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	lessonClassrooms := []*LessonClassroom{}
	for rows.Next() {
		lc := &LessonClassroom{}
		if err := rows.Scan(database.GetScanFields(lc, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonClassrooms = append(lessonClassrooms, lc)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonClassrooms, nil
}

func (l *LessonClassroomRepo) GetClassroomIDsByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonClassroomRepo.GetClassroomIDsByLessonID")
	defer span.End()

	lessonClassroom := &LessonClassroom{}
	fields, _ := lessonClassroom.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE lesson_id = $1 AND deleted_at is null`,
		strings.Join(fields, ","),
		lessonClassroom.TableName())
	rows, err := db.Query(ctx, query, lessonID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	lessonClassroomIDs := []string{}
	for rows.Next() {
		lc := &LessonClassroom{}
		if err := rows.Scan(database.GetScanFields(lc, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonClassroomIDs = append(lessonClassroomIDs, lc.ClassroomID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonClassroomIDs, nil
}

func (l *LessonClassroomRepo) GetLessonClassroomsWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]domain.LessonClassrooms, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonClassroomRepo.GetLessonClassroomsWithNamesByLessonIDs")
	defer span.End()

	lc := &LessonClassroom{}
	query := `SELECT lc.lesson_id, lc.classroom_id, c.name, c.room_area
		FROM lesson_classrooms lc 
		INNER JOIN classroom c ON c.classroom_id = lc.classroom_id
		WHERE lc.lesson_id = ANY($1) AND lc.deleted_at IS NULL AND c.deleted_at IS NULL `

	fields := []string{
		"lesson_id",
		"classroom_id",
		"name",
		"room_area",
	}

	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	// fetch results of query
	lessonClassroomsMap := make(map[string]domain.LessonClassrooms, len(lessonIDs))
	var name pgtype.Text
	var roomArea pgtype.Text

	scanFields := append(database.GetScanFields(lc, fields), &name, &roomArea)

	for rows.Next() {
		if err := rows.Scan(scanFields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonClassroom := lc.ToLessonClassroomEntity()
		lessonClassroom.WithClassroomName(name.String)
		lessonClassroom.WithClassroomArea(roomArea.String)
		lessonClassroomsMap[lc.LessonID.String] = append(lessonClassroomsMap[lc.LessonID.String], lessonClassroom)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonClassroomsMap, nil
}

func (l *LessonClassroomRepo) GetOccupiedClassroomByTime(ctx context.Context, db database.QueryExecer, locationIDs []string, lessonID string, startTime, endTime time.Time, timezone string) (*domain.LessonClassrooms, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassroomRepo.GetOccupiedClassroomByTime")
	defer span.End()

	if len(locationIDs) == 0 || startTime.IsZero() || endTime.IsZero() {
		return nil, fmt.Errorf("location_ids, start_time, end_time is required")
	}

	classroom := &LessonClassroom{}
	fields, _ := classroom.FieldMap()

	query := fmt.Sprintf(` SELECT lc.%s 
			FROM %s lc
			JOIN lessons l ON l.lesson_id = lc.lesson_id
			WHERE lc.deleted_at is null
			AND l.deleted_at is null
			AND l.center_id = any($1)
			AND not ((l.start_time at time zone $2 >= $4::timestamptz at time zone $2)
				OR ($3::timestamptz at time zone $2 >= l.end_time at time zone $2)
			)`,

		strings.Join(fields, ",lc."), classroom.TableName(),
	)
	if lessonID != "" {
		query += fmt.Sprintf(" AND l.lesson_id <> '%s'", lessonID)
	}

	rows, err := db.Query(ctx, query, locationIDs, timezone, startTime, endTime)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	classrooms := domain.LessonClassrooms{}
	for rows.Next() {
		clr := &LessonClassroom{}
		if err := rows.Scan(database.GetScanFields(clr, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		classrooms = append(classrooms, clr.ToLessonClassroomEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return &classrooms, nil
}
