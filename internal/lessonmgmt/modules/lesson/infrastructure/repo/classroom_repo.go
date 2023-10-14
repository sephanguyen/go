package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ClassroomRepo struct{}

func (c *ClassroomRepo) getClassroomsByIDs(ctx context.Context, db database.QueryExecer, classroomIDs []string) ([]*Classroom, error) {
	classroom := &Classroom{}
	fields, _ := classroom.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
						  WHERE classroom_id = ANY($1) 
						  AND deleted_at is null`,
		strings.Join(fields, ","),
		classroom.TableName(),
	)

	rows, err := db.Query(ctx, query, classroomIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	classrooms := make([]*Classroom, 0, len(classroomIDs))
	for rows.Next() {
		cr := &Classroom{}
		if err := rows.Scan(database.GetScanFields(cr, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		classrooms = append(classrooms, cr)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return classrooms, nil
}

func (c *ClassroomRepo) CheckClassroomIDs(ctx context.Context, db database.QueryExecer, classroomIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassroomRepo.CheckClassroomIDs")
	defer span.End()

	classrooms, err := c.getClassroomsByIDs(ctx, db, classroomIDs)
	if err != nil {
		return fmt.Errorf("error on fetching classrooms by IDs: %w", err)
	}

	if len(classroomIDs) != len(classrooms) {
		return fmt.Errorf("received classroom IDs %v but only found %v", classroomIDs, classrooms)
	}

	return nil
}

func (c *ClassroomRepo) ExportAllClassrooms(ctx context.Context, db database.QueryExecer, exportCols []exporter.ExportColumnMap) ([]byte, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassroomRepo.ExportAllClassrooms")
	defer span.End()

	classroom := &ClassroomToExport{}
	query := fmt.Sprintf(`SELECT clr.location_id, lc.name as location_name, clr.classroom_id, clr.name as classroom_name, 
		clr.remarks, clr.is_archived, clr.created_at, clr.updated_at, clr.deleted_at, clr.room_area, clr.seat_capacity
		FROM %s clr
		LEFT JOIN locations lc ON clr.location_id = lc.location_id
		WHERE clr.deleted_at is NULL AND lc.deleted_at is NULL
		ORDER BY 1,4`,
		classroom.TableName())
	fields, _ := classroom.FieldMap()
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all date info: db.Query")
	}
	defer rows.Close()

	allClassrooms := []*ClassroomToExport{}
	for rows.Next() {
		item := &ClassroomToExport{}
		if err := rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		allClassrooms = append(allClassrooms, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to export all classrooms rows.Err")
	}

	exportable := sliceutils.Map(allClassrooms, func(d *ClassroomToExport) database.Entity {
		return d
	})

	str, err := exporter.ExportBatch(exportable, exportCols)
	if err != nil {
		return nil, fmt.Errorf("ExportBatch: %w", err)
	}
	return exporter.ToCSV(str), nil
}

func (c *ClassroomRepo) UpsertClassrooms(ctx context.Context, db database.QueryExecer, clrs []*domain.Classroom) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassroomRepo.UpsertClassrooms")
	defer span.End()

	clrDTOs := make([]*Classroom, 0, len(clrs))
	for _, clr := range clrs {
		classroomDTO, err := NewClassrooomFromEntity(clr)
		if err != nil {
			return err
		}
		clrDTOs = append(clrDTOs, classroomDTO)
	}

	b := &pgx.Batch{}
	for _, classroom := range clrDTOs {
		fields, args := classroom.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT classroom_pk DO 
		UPDATE SET name = $2, location_id = $3, remarks = $4,
		    room_area = $9, seat_capacity = $10, updated_at = $7, deleted_at = NULL`,
			classroom.TableName(),
			strings.Join(fields, ","),
			placeHolders)
		b.Queue(query, args...)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("classrrooms not upserted")
		}
	}
	return nil
}

func (c *ClassroomRepo) RetrieveClassroomsByLocationID(ctx context.Context, db database.QueryExecer, params *payloads.GetClassroomListArg) ([]*domain.Classroom, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassroomRepo.RetrieveClassroomsByLocationID")
	defer span.End()

	classroom := &Classroom{}
	fields, _ := classroom.FieldMap()

	args := []interface{}{
		&params.LocationIDs,
	}
	paramsNum := len(args)
	baseTable := fmt.Sprintf(`select %s from %s c `, strings.Join(fields, ","), classroom.TableName())
	where := fmt.Sprintf(` WHERE c.deleted_at IS NULL AND c.location_id = any($%d)`, paramsNum)

	if strings.TrimSpace(params.KeyWord) != "" {
		where += fmt.Sprintf(` AND c.name ilike '%%%s%%' `, params.KeyWord)
	}

	query := baseTable + where + " GROUP BY c.classroom_id, c.room_area ORDER BY c.room_area ASC, c.name ASC, c.created_at ASC "
	if params.Limit > 0 {
		query += fmt.Sprintf(` LIMIT %d OFFSET %d`, params.Limit, params.Offset)
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	classrooms := []*domain.Classroom{}
	for rows.Next() {
		cr := &Classroom{}
		if err := rows.Scan(database.GetScanFields(cr, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		classrooms = append(classrooms, cr.ToClassroomEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return classrooms, nil
}
