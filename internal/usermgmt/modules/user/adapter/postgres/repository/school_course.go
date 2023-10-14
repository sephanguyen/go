package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type SchoolCourseRepo struct{}

func (r *SchoolCourseRepo) GetByIDsAndSchoolIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, schoolIDs pgtype.TextArray) ([]*entity.SchoolCourse, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolCourseRepo.GetByIDs")
	defer span.End()

	course := &entity.SchoolCourse{}
	fields := database.GetFieldNames(course)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE school_course_id = ANY($1) AND school_id = ANY($2) AND deleted_at IS NULL", strings.Join(fields, ","), course.TableName())

	rows, err := db.Query(ctx, stmt, &ids, &schoolIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]*entity.SchoolCourse, 0)
	for rows.Next() {
		course := &entity.SchoolCourse{}
		if err := rows.Scan(database.GetScanFields(course, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		courses = append(courses, course)
	}

	return courses, nil
}

func (r *SchoolCourseRepo) GetBySchoolCoursePartnerIDsAndSchoolIDs(ctx context.Context, db database.QueryExecer, schoolCoursePartnerIds pgtype.TextArray, schoolIDs pgtype.TextArray) ([]*entity.SchoolCourse, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolCourseRepo.GetByIDs")
	defer span.End()

	course := &entity.SchoolCourse{}
	courses := []*entity.SchoolCourse{}

	fields := database.GetFieldNames(course)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE school_course_partner_id = ANY($1) AND school_id = ANY($2) AND deleted_at IS NULL", strings.Join(fields, ","), course.TableName())

	rows, err := db.Query(ctx, stmt, &schoolCoursePartnerIds, &schoolIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		course := &entity.SchoolCourse{}
		if err := rows.Scan(database.GetScanFields(course, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		courses = append(courses, course)
	}

	return courses, nil
}

// Create creates SchoolCourse entity
func (r *SchoolCourseRepo) Create(ctx context.Context, db database.QueryExecer, e *entity.SchoolCourse) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolCourseRepo.Create")
	defer span.End()
	now := time.Now()

	id := e.ID.String
	if id == "" {
		id = idutil.ULIDNow()
	}

	if err := multierr.Combine(
		e.ID.Set(id),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
		e.DeletedAt.Set(nil),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set DeletedAt.Set: %w", err)
	}

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	if e.ResourcePath.Status == pgtype.Null {
		if err := e.ResourcePath.Set(resourcePath); err != nil {
			return err
		}
	}

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return fmt.Errorf("err create SchoolCourseRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err create SchoolCourseRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
