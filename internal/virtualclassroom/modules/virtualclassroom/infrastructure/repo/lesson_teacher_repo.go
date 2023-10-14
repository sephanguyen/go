package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/pkg/errors"
)

type LessonTeacherRepo struct{}

func (l *LessonTeacherRepo) GetTeacherIDsByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*LessonTeacher, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonTeacherRepo.GetTeacherIDsByLessonID")
	defer span.End()
	return l.getTeacherByLessonIDs(ctx, db, lessonIDs)
}

func (l *LessonTeacherRepo) getTeacherByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*LessonTeacher, error) {
	lessonTeacher := &LessonTeacher{}
	fields, _ := lessonTeacher.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = ANY($1) AND deleted_at is null", strings.Join(fields, ","), lessonTeacher.TableName())
	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessonTeachers := []*LessonTeacher{}
	for rows.Next() {
		lt := &LessonTeacher{}
		if err := rows.Scan(database.GetScanFields(lt, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonTeachers = append(lessonTeachers, lt)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonTeachers, nil
}

func (l *LessonTeacherRepo) GetTeachersByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]domain.LessonTeachers, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonTeacherRepo.GetTeachersByLessonIDs")
	defer span.End()
	lessonTeacher, err := l.getTeacherByLessonIDs(ctx, db, lessonIDs)
	if err != nil {
		return nil, err
	}
	lt := make(map[string]domain.LessonTeachers)
	for _, v := range lessonTeacher {
		lt[v.LessonID.String] = append(lt[v.LessonID.String],
			&domain.LessonTeacher{TeacherID: v.TeacherID.String})
	}
	return lt, nil
}

func (l *LessonTeacherRepo) GetTeacherIDsByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonTeacherRepo.GetTeacherIDsByLessonID")
	defer span.End()
	lessonTeacher := &LessonTeacher{}
	fields, _ := lessonTeacher.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = $1 AND deleted_at is null", strings.Join(fields, ","), lessonTeacher.TableName())
	rows, err := db.Query(ctx, query, lessonID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessonTeacherIDs := []string{}
	for rows.Next() {
		lt := &LessonTeacher{}
		if err := rows.Scan(database.GetScanFields(lt, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonTeacherIDs = append(lessonTeacherIDs, lt.TeacherID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonTeacherIDs, nil
}

func (l *LessonTeacherRepo) GetTeacherIDsOnlyByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonTeacherRepo.GetTeacherIDsOnlyByLessonIDs")
	defer span.End()

	lessonTeacher := &LessonTeacher{}
	fields, values := lessonTeacher.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE lesson_id = ANY($1) AND deleted_at is null`,
		strings.Join(fields, ","),
		lessonTeacher.TableName(),
	)

	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	lessonsTeachersMap := make(map[string][]string, len(lessonIDs))
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		lessonsTeachersMap[lessonTeacher.LessonID.String] = append(lessonsTeachersMap[lessonTeacher.LessonID.String], lessonTeacher.TeacherID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return lessonsTeachersMap, nil
}
