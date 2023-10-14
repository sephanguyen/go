package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eurekaDB "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type CourseStudentRepo struct {
}

func (p *CourseStudentRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.CourseStudent) error {
	now := timeutil.Now()
	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	fieldNames, values := e.FieldMap()
	placeHolders := "$1, $2, $3, $4, $5, $6"

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (course_id, student_id)
		DO UPDATE SET deleted_at = NULL, updated_at = NOW()`,
		e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	ct, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return errors.New("cannot insert course student")
	}
	return nil
}

const courseStudentRepoBulkUpsertStmtTpl = `INSERT INTO %s AS cs (%s) VALUES (%s)
ON CONFLICT (course_id, student_id)
DO UPDATE SET
	deleted_at = NULL,
	updated_at = NOW(),
	start_at = excluded.start_at,
	end_at = excluded.end_at
RETURNING course_id, student_id, course_student_id`

type CourseStudentKey struct {
	CourseID  string
	StudentID string
}

// BulkUpsert will soft-deletes rows if rows already exist
func (p *CourseStudentRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.CourseStudent) (map[CourseStudentKey]string, error) {
	b := &pgx.Batch{}
	e := &entities.CourseStudent{}
	currentTime := timeutil.Now().UTC()

	for _, item := range items {
		fieldNames, value := item.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(courseStudentRepoBulkUpsertStmtTpl, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

		if item.CreatedAt.Status != pgtype.Present && item.UpdatedAt.Status != pgtype.Present {
			b.Queue(query, append(value[:3], currentTime, currentTime, nil)...)
		} else {
			b.Queue(query, value...)
		}
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	courseStudentMap := make(map[CourseStudentKey]string)

	for i := 0; i < b.Len(); i++ {
		var courseID, studentID, courseStudentID pgtype.Text
		if err := result.QueryRow().Scan(&courseID, &studentID, &courseStudentID); err != pgx.ErrNoRows && err != nil {
			return nil, fmt.Errorf("batchResults.QueryRow: %w", err)
		}

		key := CourseStudentKey{
			CourseID:  courseID.String,
			StudentID: studentID.String,
		}
		courseStudentMap[key] = courseStudentID.String
	}
	return courseStudentMap, nil
}

const courseStudentRepoSoftDeleteStmt = `UPDATE course_students SET deleted_at = NOW()
WHERE (course_id, student_id) IN (%s)
AND deleted_at IS NULL`

func (p *CourseStudentRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageRepo.SoftDelete")
	defer span.End()

	inCondition, args := database.CompositeKeysPlaceHolders(len(studentIDs), func(i int) []interface{} {
		return []interface{}{courseIDs[i], studentIDs[i]}
	})

	_, err := db.Exec(ctx, fmt.Sprintf(courseStudentRepoSoftDeleteStmt, inCondition), args...)
	if err != nil {
		return err
	}

	return nil
}

const courseStudentRepoSoftDeleteByStudentIDStmt = `UPDATE course_students SET deleted_at = NOW()
WHERE student_id = $1
AND deleted_at IS NULL`

// SoftDeleteByStudentID deletes rows if rows not deleted yet
func (p *CourseStudentRepo) SoftDeleteByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageRepo.SoftDeleteByStudentID")
	defer span.End()

	_, err := db.Exec(ctx, courseStudentRepoSoftDeleteByStudentIDStmt, studentID)
	if err != nil {
		return err
	}

	return nil
}

const courseStudentRepoSyncBulkUpsertStmtTpl = `
INSERT INTO %s (%s) VALUES %s
ON CONFLICT (course_id, student_id)
DO UPDATE SET
	deleted_at = NULL,
	updated_at = NOW(),
	start_at = excluded.start_at,
	end_at = excluded.end_at`

const courseStudentRepoSyncSoftDeleteStmt = `
	UPDATE course_students SET deleted_at = NOW()
	WHERE student_id = ANY($1::TEXT[])
	AND deleted_at IS NULL;
`

func (p *CourseStudentRepo) SoftDeleteByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageRepo.SoftDeleteWithUpsertAction")
	defer span.End()
	_, err := db.Exec(ctx, courseStudentRepoSyncSoftDeleteStmt, studentIDs)
	if err != nil {
		return fmt.Errorf("CourseStudenRepo.SoftDeleteWithUpsertAction err: %w", err)
	}

	return nil
}

func (p *CourseStudentRepo) BulkUpsertV2(ctx context.Context, db database.QueryExecer, courseStudents []*entities.CourseStudent) error {
	err := eurekaDB.BulkUpsert(ctx, db, courseStudentRepoSyncBulkUpsertStmtTpl, courseStudents)
	if err != nil {
		return fmt.Errorf("CourseStudenRepo.BulkUpsertV2 err: %w", err)
	}

	return nil
}

type SearchStudentsFilter struct {
	CourseIDs  pgtype.TextArray
	Limit      pgtype.Int8
	Offset     pgtype.Text
	StudentIDs pgtype.TextArray
}

const searchStudentsStmt = `SELECT student_id, ARRAY_AGG(course_id) 
FROM %s 
WHERE ($1::_TEXT IS NULL OR course_id = ANY($1::_TEXT)) 
	AND ($4::_TEXT IS NULL OR student_id = ANY($4::_TEXT))
	AND ($2::TEXT IS NULL OR student_id > $2)
	AND (start_at <= now() AND now() <= end_at)
	AND deleted_at IS NULL 
GROUP BY student_id
ORDER BY student_id
LIMIT $3`

func (p *CourseStudentRepo) SearchStudents(ctx context.Context, db database.QueryExecer, filter *SearchStudentsFilter) (map[string][]string, []string, error) {
	e := &entities.CourseStudent{}
	query := fmt.Sprintf(searchStudentsStmt, e.TableName())
	rows, err := db.Query(ctx, query, filter.CourseIDs, filter.Offset, filter.Limit, filter.StudentIDs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	studentIDCourses := make(map[string][]string)
	studentIDs := make([]string, 0, filter.Limit.Int)
	for rows.Next() {
		var studentID pgtype.Text
		var courses pgtype.TextArray
		if err := rows.Scan(&studentID, &courses); err != nil {
			return nil, nil, fmt.Errorf("rows.Err: %w", err)
		}
		if studentID.Status != pgtype.Null {
			studentIDs = append(studentIDs, studentID.String)
			studentIDCourses[studentID.String] = database.FromTextArray(courses)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("rows.Err: %w", err)
	}

	return studentIDCourses, studentIDs, nil
}

func (p *CourseStudentRepo) FindStudentByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) ([]string, error) {
	courseClass := &entities.CourseStudent{}
	query := fmt.Sprintf(`SELECT student_id FROM %s WHERE deleted_at is NULL AND course_id = $1`, courseClass.TableName())
	rows, err := db.Query(ctx, query, &courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var studentIDs []string
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		studentIDs = append(studentIDs, studentID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return studentIDs, nil
}

func (p *CourseStudentRepo) GetByCourseStudents(ctx context.Context, db database.QueryExecer, courseStudents entities.CourseStudents) (entities.CourseStudents, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudentRepo.GetByCourseStudents")
	defer span.End()

	fields := database.GetFieldNames(&entities.CourseStudent{})
	studentIDCourseID := make([]string, 0, len(courseStudents))
	args := make([]interface{}, 0, len(courseStudents)*2)
	for i, v := range courseStudents {
		studentID := v.StudentID
		courseID := v.CourseID
		args = append(args, &studentID, &courseID)
		studentIDCourseID = append(studentIDCourseID, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
	}
	// placeHolder will like ($1, $2), ($3, $4), ($5, $6), ....
	placeHolder := strings.Join(studentIDCourseID, ", ")

	query := fmt.Sprintf("SELECT %s FROM course_students WHERE deleted_at IS NULL AND (student_id,course_id) IN (%s)",
		strings.Join(fields, ","), placeHolder)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(entities.CourseStudents, 0, len(courseStudents))
	for rows.Next() {
		record := entities.CourseStudent{}
		scanFields := database.GetScanFields(&record, fields)
		if err = rows.Scan(scanFields...); err != nil {
			return nil, err
		}
		res = append(res, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db.Query :%v", err)
	}

	return res, nil
}

func (p *CourseStudentRepo) RetrieveByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.CourseStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudentRepo.RetrieveByIntervalTime")
	defer span.End()
	stmt := "SELECT %s FROM %s WHERE deleted_at IS NULL AND updated_at >= ( now() - $1::interval)"
	var e entities.CourseStudent
	selectFields := database.GetFieldNames(&e)

	query := fmt.Sprintf(stmt, strings.Join(selectFields, ", "), e.TableName())

	var items entities.CourseStudents
	err := database.Select(ctx, db, query, &intervalTime).ScanAll(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (p *CourseStudentRepo) FindStudentTagByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) ([]*entities.StudentTag, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseStudentRepo.FindStudentTagByCourseID")
	defer span.End()

	var e entities.StudentTag
	var results entities.StudentTags

	stmt := `
    SELECT %s
      FROM %s ut
        JOIN tagged_user tu on ut.user_tag_id = tu.tag_id
        JOIN course_students cs on tu.user_id = cs.student_id
     WHERE ut.deleted_at IS NULL
       AND tu.deleted_at IS NULL
       AND cs.deleted_at is NULL
       AND cs.course_id = $1
     GROUP BY user_tag_id, user_tag_name
     ORDER BY ut.user_tag_name;
	`

	selectFields := database.GetFieldNames(&e)
	query := fmt.Sprintf(stmt, strings.Join(selectFields, ", "), e.TableName())

	err := database.Select(ctx, db, query, courseID).ScanAll(&results)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return results, nil
}
