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

type CourseTypeRepo struct{}

func (c *CourseTypeRepo) GetByIDs(ctx context.Context, db database.Ext, courseTypeIDs []string) ([]*domain.CourseType, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseTypeRepo.GetByIDs")
	defer span.End()
	t := &CourseType{}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE course_type_id = ANY ($1) AND deleted_at IS NULL ", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query, &courseTypeIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cTypes []*domain.CourseType
	for rows.Next() {
		cType := new(CourseType)
		if err := rows.Scan(database.GetScanFields(cType, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		cTypes = append(cTypes, cType.ToCourseTypeEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return cTypes, nil
}

// Import all if all are valid, other case revert and import nothing.
func (c *CourseTypeRepo) Import(ctx context.Context, db database.Ext, courseTypes []*domain.CourseType) error {
	ctx, span := interceptors.StartSpan(ctx, "CourseTypeRepo.Import")
	defer span.End()
	b := &pgx.Batch{}
	entity := &CourseType{}
	for _, courseType := range courseTypes {
		courseType.CreatedAt = time.Now()
		courseType.UpdatedAt = time.Now()
		fields, _ := entity.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO %s (%s) 
			VALUES (%s) ON CONFLICT(course_type_id) DO 
			UPDATE SET name = $2, is_archived = $3, remarks = $4, 
			updated_at = $5, deleted_at = NULL`, entity.TableName(), strings.Join(fields, ", "), placeHolders)
		b.Queue(query,
			&courseType.CourseTypeID,
			&courseType.Name,
			&courseType.IsArchived,
			&courseType.Remarks,
			&courseType.UpdatedAt,
			&courseType.CreatedAt,
			&courseType.DeletedAt)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		ct, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("course_type could not be upserted")
		}
	}
	return nil
}
