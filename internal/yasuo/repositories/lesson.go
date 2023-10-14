package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LessonRepo struct{}

type ControlSettingLiveLesson struct {
	Lectures                  []string `json:"lectures"`
	TeacherObservers          []string `json:"teacher_observers"`
	DefaultView               string   `json:"default_view"`
	PublishStudentVideoStatus string   `json:"publish_student_video_status"`
	UnmuteStudentAudioStatus  string   `json:"unmute_student_audio_status"`
}

func (l *LessonRepo) Create(ctx context.Context, db database.Ext, plans []*entities.Lesson) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Create")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *entities.Lesson) {
		fieldNames := database.GetFieldNames(e)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}
	var d pgtype.Timestamptz
	err := d.Set(time.Now())
	if err != nil {
		return fmt.Errorf("cannot set time lessons: %w", err)
	}

	for _, each := range plans {
		if each.LessonID.String == "" {
			return fmt.Errorf("missing lesson id")
		}
		// jpref does not send teacherID to sync, it should be validate in service layer
		//if each.TeacherID.String == "" {
		//return fmt.Errorf("missing teacher id")
		//}
		each.CreatedAt = d
		each.UpdatedAt = d
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

// BulkUpsert insert if not existed, but update when the lesson existed but have soft deleted
func (l *LessonRepo) BulkUpsert(ctx context.Context, db database.Ext, lessons []*entities.Lesson) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Create")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *entities.Lesson) {
		fieldNames := database.GetFieldNames(e)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT lessons_pk DO UPDATE SET teacher_id = $2, course_id = $3, control_settings = $4 ,created_at = $5, updated_at = $6, deleted_at = $7, end_at = $8, lesson_group_id = $9, room_id = $10, lesson_type = $11, status = $12, stream_learner_counter = $13, learner_ids = $14",
			e.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, database.GetScanFields(e, fieldNames)...)
	}

	b := &pgx.Batch{}
	var d pgtype.Timestamptz
	err := d.Set(time.Now())
	if err != nil {
		return fmt.Errorf("cannot set time lessons: %w", err)
	}

	for _, each := range lessons {
		if each.LessonID.String == "" {
			return fmt.Errorf("missing lesson id")
		}
		each.CreatedAt = d
		each.UpdatedAt = d
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

func (l *LessonRepo) FindByIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, isAll bool) (map[pgtype.Text]*entities.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindByIDs")
	defer span.End()

	e := &entities.Lesson{}
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = ANY($1)", strings.Join(fields, ","), e.TableName())
	if !isAll {
		query += " AND deleted_at IS NULL"
	}
	result := map[pgtype.Text]*entities.Lesson{}
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities.Lesson)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result[c.LessonID] = c
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return result, nil
}

func (l *LessonRepo) FindByCourseIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, isAll bool) ([]*entities.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindByIDs")
	defer span.End()

	e := &entities.Lesson{}
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE course_id = ANY($1)", strings.Join(fields, ","), e.TableName())
	if !isAll {
		query += " AND deleted_at IS NULL"
	}
	result := []*entities.Lesson{}
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities.Lesson)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return result, nil
}

func (l *LessonRepo) FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindByID")
	defer span.End()
	e := &entities.Lesson{}

	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}

func (l *LessonRepo) Update(ctx context.Context, db database.QueryExecer, lesson *entities.Lesson) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Update")
	defer span.End()
	if err := lesson.SetDefaultSchedulingStatus(); err != nil {
		return err
	}
	query := `UPDATE lessons SET updated_at = now(), teacher_id = $1, control_settings = $2, lesson_group_id = $3, course_id = $4, lesson_type = $5, "name" = $6, start_time = $7, end_time =$8, teaching_medium =$9 ,scheduling_status = $10 WHERE lesson_id = $11 AND deleted_at IS NULL`
	cmdTag, err := db.Exec(ctx, query, &lesson.TeacherID, &lesson.ControlSettings, &lesson.LessonGroupID, &lesson.CourseID, &lesson.LessonType, &lesson.Name, &lesson.StartTime, &lesson.EndTime, &lesson.TeachingMedium, &lesson.SchedulingStatus, &lesson.LessonID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot update lesson")
	}

	return nil
}

func (l *LessonRepo) SoftDelete(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.SoftDelete")
	defer span.End()

	query := "UPDATE lessons SET deleted_at = now(), updated_at = now() WHERE lesson_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &lessonIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete lesson")
	}

	return nil
}

func (l *LessonRepo) SoftDeleteByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.SoftDelete")
	defer span.End()

	query := "UPDATE lessons SET deleted_at = now(), updated_at = now() WHERE course_id = ANY($1) AND deleted_at IS NULL"
	cmdTag, err := db.Exec(ctx, query, &courseIDs)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot delete lesson")
	}

	return nil
}
func (l *LessonRepo) FindEarlierAndLatestTimeLesson(ctx context.Context, db database.Ext, courseID pgtype.Text) (*time.Time, *time.Time, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.FindByID")
	defer span.End()

	lesson := &entities.Lesson{}

	query := fmt.Sprintf("SELECT min(l.start_time), max(l.end_time) FROM %s l WHERE l.deleted_at IS NULL AND l.course_id = $1", lesson.TableName())

	var startDate, endDate *time.Time
	err := db.QueryRow(ctx, query, &courseID).Scan(&startDate, &endDate)
	if err != nil {
		return nil, nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	return startDate, endDate, nil
}

func (l *LessonRepo) UpdateRoomID(ctx context.Context, db database.QueryExecer, lessonID, roomID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateRoomID")
	defer span.End()

	query := "UPDATE lessons SET updated_at = now(), room_id = $1, status = 'LESSON_STATUS_NOT_STARTED' WHERE lesson_id = $2"
	cmdTag, err := db.Exec(ctx, query, &roomID, &lessonID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("cannot update lesson")
	}

	return nil
}

func (l *LessonRepo) CheckExisted(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (validIDs, invalidIDs []string, err error) {
	sql := `SELECT lesson_id FROM lessons WHERE lesson_id = ANY($1)`

	rows, err := db.Query(ctx, sql, &ids)
	if err != nil {
		return nil, nil, fmt.Errorf("err db.Query: %w", err)
	}

	defer rows.Close()

	mapValids := map[string]struct{}{}

	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return nil, nil, fmt.Errorf("rows.Scan: %w", err)
		}

		validIDs = append(validIDs, id.String)
		mapValids[id.String] = struct{}{}
	}

	for _, id := range ids.Elements {
		if _, ok := mapValids[id.String]; !ok {
			invalidIDs = append(invalidIDs, id.String)
		}
	}

	return validIDs, invalidIDs, nil
}

func (l *LessonRepo) GetLiveLessons(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (validIDs []string, err error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLiveLessons")
	defer span.End()
	sql := `SELECT lesson_id FROM lessons WHERE lesson_id = ANY($1) AND lesson_type = 'LESSON_TYPE_ONLINE' `
	rows, err := db.Query(ctx, sql, &ids)
	if err != nil {
		return nil, fmt.Errorf("err db.Query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		validIDs = append(validIDs, id.String)
	}

	return validIDs, nil
}
