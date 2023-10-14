package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type NotificationStudentCourseRepo struct{}

type FindNotificationStudentCourseFilter struct {
	StudentCourseID pgtype.Text
	StudentID       pgtype.Text
	CourseID        pgtype.Text
	LocationID      pgtype.Text
}

func (f *FindNotificationStudentCourseFilter) Validate() error {
	if f.StudentCourseID.Status == pgtype.Null {
		if f.StudentID.Status == pgtype.Null || f.CourseID.Status == pgtype.Null {
			return fmt.Errorf("FindNotificationStudentCourseFilter student_course_id is not null or the rest is not null")
		}
	}
	return nil
}

func NewFindNotificationStudentCourseFilter() *FindNotificationStudentCourseFilter {
	f := &FindNotificationStudentCourseFilter{}
	_ = f.StudentCourseID.Set(nil)
	_ = f.StudentID.Set(nil)
	_ = f.CourseID.Set(nil)
	_ = f.LocationID.Set(nil)
	return f
}

func (n *NotificationStudentCourseRepo) Find(ctx context.Context, db database.QueryExecer, filter *FindNotificationStudentCourseFilter) (entities.NotificationStudentCourses, error) {
	ctx, span := interceptors.StartSpan(ctx, "NotificationStudentCourseRepo.Find")
	defer span.End()

	if err := filter.Validate(); err != nil {
		return nil, err
	}
	e := &entities.NotificationStudentCourse{}
	fields := strings.Join(database.GetFieldNames(e), ", ")

	query := fmt.Sprintf(`
		SELECT %s 
		FROM notification_student_courses
		WHERE ($1::TEXT IS NULL OR student_course_id = $1::TEXT)
			AND ($2::TEXT IS NULL OR student_id = $2::TEXT)
			AND ($3::TEXT IS NULL OR course_id = $3::TEXT)
			AND ($4::TEXT IS NULL OR location_id = $4::TEXT)
			AND deleted_at IS NULL;
	`, fields)

	res := entities.NotificationStudentCourses{}
	err := database.Select(ctx, db, query, filter.StudentCourseID, filter.StudentID, filter.CourseID, filter.LocationID).ScanAll(&res)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (n *NotificationStudentCourseRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.NotificationStudentCourse) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationStudentCourseRepo.Upsert")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		e.UpdatedAt.Set(now),
	)

	if e.CreatedAt.Status != pgtype.Present || e.CreatedAt.Time.IsZero() {
		err = multierr.Append(err, e.CreatedAt.Set(now))
	}
	if e.StudentCourseID.Status != pgtype.Present || e.StudentCourseID.String == "" {
		err = multierr.Append(err, e.StudentCourseID.Set(idutil.ULIDNow()))
	}
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fieldNames := database.GetFieldNames(e)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__notification_student_courses
		DO UPDATE SET 
			location_id = $4, 
			start_at = $5, 
			end_at = $6, 
			updated_at = $8, 
			deleted_at = $9
	`, e.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := database.GetScanFields(e, fieldNames)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return errors.Wrap(err, "r.DB.ExecEx")
	}
	return nil
}

func (n *NotificationStudentCourseRepo) queueCreate(b *pgx.Batch, item *entities.NotificationStudentCourse) error {
	now := time.Now()
	err := multierr.Combine(
		item.UpdatedAt.Set(now),
		item.CreatedAt.Set(now),
	)

	if item.StudentCourseID.Status != pgtype.Present || item.StudentCourseID.String == "" {
		err = multierr.Append(err, item.StudentCourseID.Set(idutil.ULIDNow()))
	}
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)
	return nil
}

func (n *NotificationStudentCourseRepo) BulkCreate(ctx context.Context, db database.QueryExecer, items []*entities.NotificationStudentCourse) error {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationAccessPathRepo.BulkCreate")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		err := n.queueCreate(b, item)
		if err != nil {
			return err
		}
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

func (n *NotificationStudentCourseRepo) queueUpsert(b *pgx.Batch, item *entities.NotificationStudentCourse) {
	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__notification_student_courses
		DO UPDATE SET location_id = $4, start_at = $5, end_at = $6, updated_at = $8, deleted_at = $9 
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)
}

func (n *NotificationStudentCourseRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.NotificationStudentCourse) error {
	ctx, span := interceptors.StartSpan(ctx, "InfoNotificationAccessPathRepo.BulkUpsert")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range items {
		n.queueUpsert(b, item)
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

type SoftDeleteNotificationStudentCourseFilter struct {
	StudentCourseIDs pgtype.TextArray
	StudentIDs       pgtype.TextArray
	CourseIDs        pgtype.TextArray
	LocationIDs      pgtype.TextArray
}

func (f *SoftDeleteNotificationStudentCourseFilter) Validate() error {
	if f.StudentCourseIDs.Status == pgtype.Null &&
		f.StudentIDs.Status == pgtype.Null {
		return fmt.Errorf("FindNotificationStudentCourseFilter: student_course_ids filter and student_ids filter are null (violate)")
	}

	return nil
}

func NewSoftDeleteNotificationStudentCourseFilter() *SoftDeleteNotificationStudentCourseFilter {
	f := &SoftDeleteNotificationStudentCourseFilter{}
	_ = f.StudentCourseIDs.Set(nil)
	_ = f.StudentIDs.Set(nil)
	_ = f.CourseIDs.Set(nil)
	_ = f.LocationIDs.Set(nil)
	return f
}

func (n *NotificationStudentCourseRepo) SoftDelete(ctx context.Context, db database.QueryExecer, filter *SoftDeleteNotificationStudentCourseFilter) error {
	ctx, span := interceptors.StartSpan(ctx, "NotificationStudentCourseRepo.SoftDelete")
	defer span.End()

	if err := filter.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE notification_student_courses 
		SET deleted_at = NOW() 
		WHERE ($1::TEXT[] IS NULL OR student_course_id = ANY($1))
			AND ($2::TEXT[] IS NULL OR student_id = ANY($2))
			AND ($3::TEXT[] IS NULL OR course_id = ANY($3))
			AND ($4::TEXT[] IS NULL OR location_id = ANY($4))
			AND deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, filter.StudentCourseIDs, filter.StudentIDs, filter.CourseIDs, filter.LocationIDs)

	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}
