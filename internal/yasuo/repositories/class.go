package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type ClassRepo struct{}

func (r *ClassRepo) RetrieveClassIDsBySchoolIDs(ctx context.Context, db database.QueryExecer, schoolIDs []int) (map[int][]int, error) {
	var data []struct {
		ClassID  int
		SchoolID int
	}
	query := `SELECT class_id, school_id FROM classes 
		WHERE school_id = ANY($1) AND status = 'CLASS_STATUS_ACTIVE'`
	rows, err := db.Query(ctx, query, schoolIDs)
	if err != nil {
		return nil, fmt.Errorf("RetrieveClassIDsBySchoolIDs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var classID, schoolID int
		err := rows.Scan(&classID, &schoolID)
		if err != nil {
			return nil, err
		}
		data = append(data, struct {
			ClassID  int
			SchoolID int
		}{
			ClassID:  classID,
			SchoolID: schoolID,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	m := make(map[int][]int) // school id => list of class ids
	for _, d := range data {
		m[d.SchoolID] = append(m[d.SchoolID], d.ClassID)
	}
	return m, nil
}

func (r *ClassRepo) FindBySchool(ctx context.Context, db database.QueryExecer, schoolID int32) ([]*entities.Class, error) {
	var p entities.Classes

	en := &entities.Class{}
	fieldNames := database.GetFieldNames(en)
	stmt := "SELECT %s FROM %s WHERE school_id = $1"
	query := fmt.Sprintf(stmt, strings.Join(fieldNames, ", "), en.TableName())
	err := database.Select(ctx, db, query, schoolID).ScanAll(&p)
	if err != nil {
		return nil, fmt.Errorf("FindBySchool: %w", err)
	}
	return p, nil
}

func (r *ClassRepo) FindBySchoolAndID(ctx context.Context, db database.QueryExecer, schoolID pgtype.Int4, classIDs pgtype.Int4Array) (map[pgtype.Int4]*entities.Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.FindBySchoolAndID")
	defer span.End()

	e := new(entities.Class)
	fields := database.GetFieldNames(e)

	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE status = 'CLASS_STATUS_ACTIVE' AND deleted_at IS NULL AND school_id = $1 AND class_id = ANY($2)", strings.Join(fields, ","), e.TableName())
	classes := map[pgtype.Int4]*entities.Class{}
	rows, err := db.Query(ctx, selectStmt, &schoolID, &classIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		c := new(entities.Class)
		if err := rows.Scan(database.GetScanFields(c, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		classes[c.ID] = c
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return classes, nil
}
