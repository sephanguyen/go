package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type ClassStudentRepo struct {
}

const classStudentRepoUpsertStmt = `INSERT INTO class_students AS cs (%s) VALUES (%s)
ON CONFLICT (student_id, class_id)
DO UPDATE SET
	deleted_at = NULL,
	updated_at = NOW()
WHERE cs.deleted_at IS NOT NULL`

// Upsert but do mainly delete... WTF???
func (r *ClassStudentRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.ClassStudent) error {
	now := timeutil.Now()
	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	fieldNames, values := e.FieldMap()
	placeHolders := "$1, $2, $3, $4, $5"

	query := fmt.Sprintf(classStudentRepoUpsertStmt, strings.Join(fieldNames, ","), placeHolders)

	ct, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return errors.New("cannot upsert class student")
	}
	return nil
}

func (r *ClassStudentRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs, classIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.SoftDelete")
	defer span.End()

	entity := &entities.ClassStudent{}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE student_id = ANY($1) AND class_id = ANY($2) AND deleted_at IS NULL`, entity.TableName())

	_, err := db.Exec(ctx, query, studentIDs, classIDs)
	if err != nil {
		return err
	}
	return nil
}

func (r *ClassStudentRepo) BulkSoftDelete(ctx context.Context, db database.QueryExecer, classStudents entities.ClassStudents) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.BulkSoftDelete")
	defer span.End()

	b := &pgx.Batch{}

	entity := &entities.ClassStudent{}
	const softDeleteClassStudentTmpl = "UPDATE %s SET deleted_at = NOW() WHERE class_id = $1 AND student_id = $2 AND deleted_at IS NULL"
	query := fmt.Sprintf(softDeleteClassStudentTmpl, entity.TableName())

	for _, classStudent := range classStudents {
		b.Queue(query, classStudent.ClassID, classStudent.StudentID)
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

func (r *ClassStudentRepo) SoftDeleteByCourseStudent(ctx context.Context, db database.QueryExecer, courseStudent *entities.CourseStudent) error {
	const softDeleteClassStudentTmpl = `
UPDATE %s AS cs
SET deleted_at = NOW()
FROM course_classes cc
WHERE cc.class_id = cs.class_id
  AND cc.course_id = $1
  AND cs.student_id = $2
  AND cs.deleted_at IS NULL
  `
	query := fmt.Sprintf(softDeleteClassStudentTmpl, (&entities.ClassStudent{}).TableName())
	_, err := db.Exec(ctx, query, courseStudent.CourseID, courseStudent.StudentID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *ClassStudentRepo) GetClassStudentByCourseAndClassIds(ctx context.Context, db database.QueryExecer, courseIDs, classIDs pgtype.TextArray) ([]*entities.ClassStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassStudentRepo.GetClassStudentByCourseAndClassIds")
	defer span.End()

	classStudent := &entities.ClassStudent{}
	courseClass := &entities.CourseClass{}
	classStudents := &entities.ClassStudents{}

	values, _ := classStudent.FieldMap()
	stmt := fmt.Sprintf(`
	SELECT cs.%s
	FROM %s cs 
	INNER JOIN %s cc
	ON cs.class_id = cc.class_id
	WHERE cs.deleted_at IS NULL
	AND cc.class_id = ANY($1::_TEXT)
	AND cc.course_id = ANY($2::_TEXT);`, strings.Join(values, ", cs."), classStudent.TableName(), courseClass.TableName())

	if err := database.Select(ctx, db, stmt, classIDs, courseIDs).ScanAll(classStudents); err != nil {
		return nil, err
	}

	return *classStudents, nil
}

func (r *ClassStudentRepo) GetClassStudentByCourse(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.ClassStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassStudentRepo.GetClassStudentByCourse")
	defer span.End()

	classStudent := &entities.ClassStudent{}
	courseClass := &entities.CourseClass{}
	classStudents := &entities.ClassStudents{}

	values, _ := classStudent.FieldMap()
	stmt := fmt.Sprintf(`
	SELECT cs.%s
	FROM %s cs 
	INNER JOIN %s cc
	ON cs.class_id = cc.class_id
	WHERE cs.deleted_at IS NULL
	AND cc.course_id = ANY($1::_TEXT);`, strings.Join(values, ", cs."), classStudent.TableName(), courseClass.TableName())

	if err := database.Select(ctx, db, stmt, courseIDs).ScanAll(classStudents); err != nil {
		return nil, err
	}

	return *classStudents, nil
}
