package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type CourseRepo struct{}

func (c *CourseRepo) UpdateTeachingMethod(ctx context.Context, db database.Ext, courseList []*domain.Course) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.UpdateTeachingMethod")
	defer span.End()

	b := &pgx.Batch{}
	for _, course := range courseList {
		courseID := database.Text(course.CourseID)
		teachingMethod := database.Text(string(course.TeachingMethod))
		b.Queue(`UPDATE courses SET teaching_method = $1 WHERE course_id = $2`, teachingMethod, courseID)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()
	// check for data after batch updating
	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (c *CourseRepo) Upsert(ctx context.Context, db database.Ext, cc []*domain.Course) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, c *Course) {
		fieldNames := database.GetFieldNames(c)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT courses_pk DO UPDATE
		SET updated_at = now(), deleted_at = NULL, name = $2, icon = $8, teaching_method = $9, course_type_id = $10`,
			c.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(c, fieldNames)...)
	}
	b := &pgx.Batch{}

	for _, c := range cc {
		course, err := NewCourseFromEntity(c)
		if err != nil {
			return fmt.Errorf("NewCourseFromEntity err: %w", err)
		}
		queue(b, course)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course not inserted")
		}
	}
	return nil
}

func (c *CourseRepo) LinkSubjects(ctx context.Context, db database.Ext, courses []*domain.Course) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.LinkSubjects")
	defer span.End()

	queue := func(b *pgx.Batch, c *domain.Course) {
		courseSubject := &CourseSubject{
			CourseID: database.Varchar(c.CourseID),
		}
		deleteQ := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE course_id = $1`,
			courseSubject.TableName())
		b.Queue(deleteQ, c.CourseID)

		for _, subjectID := range c.SubjectIDs {
			courseSubject.SubjectID = database.Varchar(subjectID)

			upsertQ := fmt.Sprintf(`INSERT INTO %s (course_id, subject_id, created_at, updated_at)
			VALUES ($1, $2, now(), now()) 
			ON CONFLICT ON CONSTRAINT course_subject_pkey 
			DO UPDATE 
			SET deleted_at = NULL, updated_at = now()`,
				courseSubject.TableName())
			b.Queue(upsertQ,
				courseSubject.CourseID.String,
				courseSubject.SubjectID.String)
		}
	}
	batch := &pgx.Batch{}

	for _, c := range courses {
		queue(batch, c)
	}
	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(courses); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (c *CourseRepo) GetByID(ctx context.Context, db database.QueryExecer, courseID string) (*domain.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.GetByID")
	defer span.End()

	e := &Course{}
	fields, args := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE course_id = $1 AND deleted_at IS NULL", strings.Join(fields, ", "), e.TableName())
	err := db.QueryRow(ctx, query, &courseID).Scan(args...)
	if err != nil {
		return nil, err
	}
	return e.ToCourseEntity(), nil
}

func (c *CourseRepo) GetByIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) ([]*domain.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.GetByID")
	defer span.End()

	dto := &Course{}
	fields, _ := dto.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE course_id = ANY($1)
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		dto.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(courseIDs))
	if err != nil {
		return nil, err
	}

	return readCourse(rows)
}

func (c *CourseRepo) GetAll(ctx context.Context, db database.QueryExecer) ([]*Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.GetAll")
	defer span.End()

	dto := &Course{}
	fields, _ := dto.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE deleted_at IS NULL`,
		strings.Join(fields, ","),
		dto.TableName(),
	)

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	cs := []*Course{}
	defer rows.Close()
	for rows.Next() {
		cf := new(Course)
		if err := rows.Scan(database.GetScanFields(cf, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		cf.TeachingMethod = database.Text(domain.ConvertTeachingMethodToString(cf.TeachingMethod.String))
		cs = append(cs, cf)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return cs, nil
}

// Import all if all are valid, other case revert and import nothing.
func (c *CourseRepo) Import(ctx context.Context, db database.Ext, courses []*domain.Course) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.Import")
	defer span.End()
	b := &pgx.Batch{}
	for _, course := range courses {
		course.CreatedAt = time.Now()
		course.UpdatedAt = time.Now()
		dto, _ := NewCourseFromEntity(course)
		fields := []string{
			"course_id",
			"name",
			"course_type_id",
			"is_archived",
			"remarks",
			"course_partner_id",
			"updated_at",
			"created_at",
			"school_id",
			"teaching_method",
		}
		placeHolders := database.GeneratePlaceholders(len(fields))

		query := fmt.Sprintf("INSERT INTO courses (%s) "+
			"VALUES (%s) ON CONFLICT(course_id) DO "+
			"UPDATE SET name = $2, course_type_id = $3, is_archived = $4, remarks = $5, "+
			"course_partner_id = $6, updated_at = $7, deleted_at = NULL, school_id = $9, teaching_method = $10",
			strings.Join(fields, ", "), placeHolders)
		b.Queue(query,
			&dto.ID,
			&dto.Name,
			&dto.CourseTypeID,
			&dto.IsArchived,
			&dto.Remarks,
			&dto.PartnerID,
			&dto.UpdatedAt,
			&dto.CreatedAt,
			&dto.SchoolID,
			&dto.TeachingMethod,
		)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		ct, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course is not upserted")
		}
	}
	return nil
}

func (c *CourseRepo) GetByPartnerIDs(ctx context.Context, db database.QueryExecer, partnerIDs []string) ([]*domain.Course, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.GetByPartnerIDs")
	defer span.End()

	dto := &Course{}
	fields, _ := dto.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE course_partner_id = ANY($1)
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		dto.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(partnerIDs))
	if err != nil {
		return nil, err
	}

	return readCourse(rows)
}

func readCourse(rows pgx.Rows) ([]*domain.Course, error) {
	var cfs []*domain.Course
	dto := &Course{}
	fields, _ := dto.FieldMap()

	defer rows.Close()
	for rows.Next() {
		cf := new(Course)
		if err := rows.Scan(database.GetScanFields(cf, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		cfs = append(cfs, cf.ToCourseEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return cfs, nil
}
