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
	"github.com/jackc/pgx/v4"
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
			&domain.LessonTeacher{
				TeacherID: v.TeacherID.String,
				Name:      v.TeacherName.String,
			})
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

func (l *LessonTeacherRepo) UpdateLessonTeacherNames(ctx context.Context, db database.QueryExecer, lessonTeachers []*domain.UpdateLessonTeacherName) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonTeacherRepo.UpdateLessonTeacherNames")
	defer span.End()
	b := &pgx.Batch{}
	strQuery := `UPDATE lessons_teachers  
		SET teacher_name = $2, updated_at =$3
		WHERE teacher_id = $1  `

	for _, lessonTeacher := range lessonTeachers {
		b.Queue(strQuery, lessonTeacher.TeacherID, lessonTeacher.FullName, time.Now())
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

func (l *LessonTeacherRepo) GetTeachersWithNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, _ bool) (map[string]domain.LessonTeachers, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonTeacherRepo.GetTeachersWithNamesByLessonIDs")
	defer span.End()

	lessonTeacher := &LessonTeacher{}
	baseQuery := ` SELECT lt.lesson_id, lt.teacher_id `
	whereClause := ` WHERE lt.lesson_id = ANY($1) AND lt.deleted_at is null `

	baseQuery += ` ,ubi.name FROM lessons_teachers lt
			JOIN user_basic_info ubi ON ubi.user_id = lt.teacher_id `
	whereClause += ` AND ubi.deleted_at is null `

	fields := []string{
		"lesson_id",
		"teacher_id",
	}
	query := baseQuery + whereClause

	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	// fetch results of query
	lessonTeachersMap := make(map[string]domain.LessonTeachers, len(lessonIDs))
	var name pgtype.Text
	scanFields := append(database.GetScanFields(lessonTeacher, fields), &name)

	for rows.Next() {
		if err := rows.Scan(scanFields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		lessonTeachersMap[lessonTeacher.LessonID.String] = append(lessonTeachersMap[lessonTeacher.LessonID.String],
			&domain.LessonTeacher{
				TeacherID: lessonTeacher.TeacherID.String,
				Name:      name.String,
			})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonTeachersMap, nil
}
