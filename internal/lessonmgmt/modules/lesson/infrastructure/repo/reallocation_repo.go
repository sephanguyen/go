package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ReallocationRepo struct{}

func (l *ReallocationRepo) UpsertReallocation(ctx context.Context, db database.QueryExecer, lessonID string, reallocations []*domain.Reallocation) error {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.UpsertReallocation")
	defer span.End()
	reallocationEntity, err := NewReallocateStudentFromEntity(reallocations)
	if err != nil {
		return err
	}
	return l.upsertReallocation(ctx, db, lessonID, reallocationEntity)
}

func (l *ReallocationRepo) upsertReallocation(ctx context.Context, db database.QueryExecer, lessonID string, reallocations []*Reallocation) error {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.upsertReallocation")
	defer span.End()

	b := &pgx.Batch{}
	for i := range reallocations {
		if err := reallocations[i].PreUpsert(); err != nil {
			return fmt.Errorf("could not pre-upsert reallocation %s", reallocations[i].StudentID.String)
		}
		upsertFields := []string{
			"student_id",
			"original_lesson_id",
			"new_lesson_id",
			"course_id",
			"updated_at",
			"created_at",
		}
		l.queueUpsertReallocateStudent(b, reallocations[i], upsertFields)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return nil
}

func (l *ReallocationRepo) queueUpsertReallocateStudent(b *pgx.Batch, r *Reallocation, upsertFields []string) {
	args := database.GetScanFields(r, upsertFields)
	placeHolders := database.GeneratePlaceholders(len(upsertFields))
	updatePlaceHolders := database.GenerateUpdatePlaceholders(upsertFields, 1)
	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT reallocation__pk DO "+
		"UPDATE SET deleted_at = NULL, %s ",
		r.TableName(),
		strings.Join(upsertFields, ", "),
		placeHolders,
		updatePlaceHolders,
	)
	b.Queue(sql, args...)
}

func (l *ReallocationRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentOriginalLesson []string, isReallocated bool) error {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.SoftDelete")
	defer span.End()

	if len(studentOriginalLesson)%2 != 0 {
		return fmt.Errorf("student of lessons are invalid")
	}

	query := fmt.Sprintf(`UPDATE reallocation SET deleted_at = now(), updated_at = now() 
						 WHERE (student_id,original_lesson_id) IN (:PlaceHolderVar) `)
	if !isReallocated {
		query += "and new_lesson_id is null"
	}
	args := []interface{}{}
	studentIDWithLessonID := make([]string, 0, len(studentOriginalLesson)/2) // will like ["($1, $2)", "($3, $4)", ...]
	for i := 0; i < len(studentOriginalLesson); i += 2 {
		studentId := studentOriginalLesson[i]
		lessonId := studentOriginalLesson[i+1]
		args = append(args, &studentId, &lessonId)
		studentIDWithLessonID = append(studentIDWithLessonID, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}
	// placeHolderVar will like ($1, $2), ($3, $4), ($5, $6), ....
	placeHolderVar := strings.Join(studentIDWithLessonID, ", ")
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)
	_, err := db.Exec(ctx, query, args...)
	return err
}

func (l *ReallocationRepo) GetFollowingReallocation(ctx context.Context, db database.QueryExecer, originalLesson string, studentId []string) ([]*domain.Reallocation, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.BreakFollowingReallocation")
	defer span.End()

	query := fmt.Sprintf(` WITH RECURSIVE al AS ( SELECT student_id,original_lesson_id,new_lesson_id  FROM reallocation
		WHERE  original_lesson_id = $1 and student_id = ANY($2)
		UNION SELECT r.student_id,r.original_lesson_id,r.new_lesson_id
		FROM reallocation AS r JOIN al ON al.new_lesson_id = r.original_lesson_id) SELECT *  FROM al   `)

	rows, err := db.Query(ctx, query, &originalLesson, &studentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*domain.Reallocation
	fields := []string{"student_id", "original_lesson_id", "new_lesson_id"}
	for rows.Next() {
		r := &Reallocation{}
		if err := rows.Scan(database.GetScanFields(r, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		res = append(res, &domain.Reallocation{
			OriginalLessonID: r.OriginalLessonID.String,
			StudentID:        r.StudentID.String,
			NewLessonID:      r.NewLessonID.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return res, err
}

func (l *ReallocationRepo) CancelIfStudentReallocated(ctx context.Context, db database.QueryExecer, studentOfNewLesson []string) error {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.CancelIfStudentReallocated")
	defer span.End()

	if len(studentOfNewLesson)%2 != 0 {
		return fmt.Errorf("student of lessons are invalid")
	}

	query := `UPDATE reallocation r SET new_lesson_id = NULL, updated_at = now() WHERE (student_id,new_lesson_id) IN (:PlaceHolderVar) 
						AND not exists (select 1 from reallocation r2 where r2.original_lesson_id = r.new_lesson_id and r2.deleted_at is null)`
	args := []interface{}{}
	studentIDWithLessonID := make([]string, 0, len(studentOfNewLesson)/2) // will like ["($1, $2)", "($3, $4)", ...]
	for i := 0; i < len(studentOfNewLesson); i += 2 {
		studentId := studentOfNewLesson[i]
		lessonId := studentOfNewLesson[i+1]
		args = append(args, &studentId, &lessonId)
		studentIDWithLessonID = append(studentIDWithLessonID, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}
	// placeHolderVar will like ($1, $2), ($3, $4), ($5, $6), ....
	placeHolderVar := strings.Join(studentIDWithLessonID, ", ")
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)
	_, err := db.Exec(ctx, query, args...)
	return err
}

func (l *ReallocationRepo) GetReallocatedLesson(ctx context.Context, db database.QueryExecer, lessonMembers []string) ([]*domain.Reallocation, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.GetReallocatedLessonID")
	defer span.End()

	if len(lessonMembers)%2 != 0 {
		return nil, fmt.Errorf("invalid lesson member input value,got %s", lessonMembers)
	}
	lessonMember := make([]string, 0, len(lessonMembers)/2) // will like ["($1, $2)", "($3, $4)", ...]
	args := make([]interface{}, 0, len(lessonMembers))
	for i := 0; i < len(lessonMembers); i += 2 {
		lessonID := lessonMembers[i]
		studentID := lessonMembers[i+1]
		args = append(args, lessonID, studentID)
		lessonMember = append(lessonMember, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}
	reallocation := &Reallocation{}
	fields, _ := reallocation.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE (original_lesson_id,student_id) IN (:PlaceHolderVar) AND deleted_at is null", strings.Join(fields, ", "), reallocation.TableName())
	placeHolderVar := strings.Join(lessonMember, ", ")
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	var res []*domain.Reallocation
	for rows.Next() {
		r := &Reallocation{}
		if err := rows.Scan(database.GetScanFields(r, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		res = append(res, &domain.Reallocation{
			OriginalLessonID: r.OriginalLessonID.String,
			StudentID:        r.StudentID.String,
			NewLessonID:      r.NewLessonID.String,
			CourseID:         r.CourseID.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return res, nil
}

func (l *ReallocationRepo) GetByNewLessonIDAndStudentID(ctx context.Context, db database.QueryExecer, lessonMembers []string) ([]*domain.Reallocation, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.GetReallocatedLessonID")
	defer span.End()
	var res []*domain.Reallocation

	if len(lessonMembers) == 0 {
		return res, nil
	}

	if len(lessonMembers)%2 != 0 {
		return nil, fmt.Errorf("invalid lesson member input value,got %s", lessonMembers)
	}
	lessonMember := make([]string, 0, len(lessonMembers)/2) // will like ["($1, $2)", "($3, $4)", ...]
	args := make([]interface{}, 0, len(lessonMembers))
	for i := 0; i < len(lessonMembers); i += 2 {
		lessonID := lessonMembers[i]
		studentID := lessonMembers[i+1]
		args = append(args, lessonID, studentID)
		lessonMember = append(lessonMember, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}
	reallocation := &Reallocation{}
	fields, _ := reallocation.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE (new_lesson_id,student_id) IN (:PlaceHolderVar) AND deleted_at is null", strings.Join(fields, ", "), reallocation.TableName())
	placeHolderVar := strings.Join(lessonMember, ", ")
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := &Reallocation{}
		if err := rows.Scan(database.GetScanFields(r, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		res = append(res, &domain.Reallocation{
			OriginalLessonID: r.OriginalLessonID.String,
			StudentID:        r.StudentID.String,
			NewLessonID:      r.NewLessonID.String,
			CourseID:         r.CourseID.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return res, nil
}

func (l *ReallocationRepo) DeleteByOriginalLessonID(ctx context.Context, db database.QueryExecer, originalLesson []string) error {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.DeleteByOriginalLessonID")
	defer span.End()
	sql := `UPDATE reallocation
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE original_lesson_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &originalLesson)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

func (l *ReallocationRepo) CancelReallocationByLessonID(ctx context.Context, db database.QueryExecer, newLessonID []string) error {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.CancelReallocationByLessonID")
	defer span.End()

	query := `UPDATE reallocation r SET new_lesson_id = NULL, updated_at = now() 
						  WHERE r.new_lesson_id = ANY($1) and r.deleted_at is null
						  AND not exists (select 1 from reallocation r2 where r2.original_lesson_id = r.new_lesson_id and r2.deleted_at is null)`
	_, err := db.Exec(ctx, query, &newLessonID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

func (l *ReallocationRepo) GetByNewLessonID(ctx context.Context, db database.QueryExecer, studentID []string, newLessonID string) ([]*domain.Reallocation, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.GetReallocatedLessonID")
	defer span.End()
	real := &Reallocation{}
	fields, _ := (&Reallocation{}).FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = ANY($1) AND new_lesson_id = $2 AND deleted_at is null", strings.Join(fields, ", "), real.TableName())

	rows, err := db.Query(ctx, query, studentID, newLessonID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	var res []*domain.Reallocation
	for rows.Next() {
		r := &Reallocation{}
		if err := rows.Scan(database.GetScanFields(r, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		res = append(res, &domain.Reallocation{
			StudentID:        r.StudentID.String,
			NewLessonID:      r.NewLessonID.String,
			OriginalLessonID: r.OriginalLessonID.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return res, nil
}
