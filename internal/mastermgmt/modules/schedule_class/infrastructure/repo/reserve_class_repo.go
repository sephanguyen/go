package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ReserveClassRepo struct{}

func (rc *ReserveClassRepo) InsertOne(ctx context.Context, db database.QueryExecer, reserveClassDomain *domain.ReserveClass) error {
	ctx, span := interceptors.StartSpan(ctx, "ReserveClassRepo.InsertOne")
	defer span.End()

	reserveClass, err := NewReserveClassFromEntity(reserveClassDomain)
	if err != nil {
		return err
	}
	fields, value := reserveClass.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		reserveClass.TableName(),
		strings.Join(fields, ","),
		placeHolders)

	_, err = db.Exec(ctx, query, value...)
	return err
}

func (rc *ReserveClassRepo) DeleteOldReserveClass(ctx context.Context, db database.QueryExecer, studentPackageID, studentID, courseID string) (pgtype.Text, pgtype.Date, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReserveClassRepo.DeleteOldReserveClass")
	defer span.End()

	strQuery := `update reserve_class rc SET deleted_at = NOW() 
			WHERE student_package_id = $1 and student_id = $2 and course_id = $3 and deleted_at is null
			RETURNING rc.class_id, rc.effective_date`

	var classID pgtype.Text
	var effectiveDate pgtype.Date

	err := db.QueryRow(ctx, strQuery, studentPackageID, studentID, courseID).Scan(&classID, &effectiveDate)

	if err != nil && err != pgx.ErrNoRows {
		return classID, effectiveDate, fmt.Errorf("failed to scan old reserve class info on %s and %v: %w", classID.String, effectiveDate.Time, err)
	}

	return classID, effectiveDate, nil
}

func (rc *ReserveClassRepo) GetByStudentIDs(ctx context.Context, db database.QueryExecer, studentID string) ([]*domain.ReserveClass, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReserveClassRepo.GetByStudentIDs")
	defer span.End()
	reserveClassDTO := &ReserveClassDTO{}
	fields := database.GetFieldNames(reserveClassDTO)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND effective_date > now() AND deleted_at IS NULL", strings.Join(fields, ","), reserveClassDTO.TableName())
	rows, err := db.Query(ctx, query, &studentID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var itemList []*domain.ReserveClass
	for rows.Next() {
		item := new(ReserveClassDTO)
		if err := rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		itemList = append(itemList, item.ToReserveClassDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return itemList, nil
}

func (rc *ReserveClassRepo) GetByEffectiveDate(ctx context.Context, db database.QueryExecer, date string) ([]*domain.ReserveClass, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReserveClassRepo.GetByStudentIDs")
	defer span.End()
	reserveClassDTO := &ReserveClassDTO{}
	fields := database.GetFieldNames(reserveClassDTO)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL and effective_date::date = '%s'", strings.Join(fields, ","), reserveClassDTO.TableName(), date)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	itemList := make([]*domain.ReserveClass, 0)
	for rows.Next() {
		item := new(ReserveClassDTO)
		if err := rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		itemList = append(itemList, item.ToReserveClassDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return itemList, nil
}

func (rc *ReserveClassRepo) DeleteByEffectiveDate(ctx context.Context, db database.QueryExecer, date string) error {
	reserveClassDTO := &ReserveClassDTO{}

	stmt := fmt.Sprintf(`UPDATE %s 
		SET deleted_at = NOW()
		WHERE deleted_at IS NULL and effective_date::date = '%s'`, reserveClassDTO.TableName(), date)
	cmd, err := db.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any reserve class to delete: %w", pgx.ErrNoRows)
	}

	return nil
}
