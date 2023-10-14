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
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type SchoolInfoRepo struct{}

type ImportError struct {
	RowNumber int32
	Error     string
}

// Import SchoolInfo
func (r *SchoolInfoRepo) BulkImport(ctx context.Context, db database.QueryExecer, items []*entity.SchoolInfo) (errors []*ImportError) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolInfoRepo.BulkImport")
	defer span.End()
	b := &pgx.Batch{}
	mappers := make(map[int]int)
	i := 0
	for order, schoolInfo := range items {
		mappers[i] = order
		i++
		fields, args := schoolInfo.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf("INSERT INTO school_info (%s) "+
			"VALUES (%s) ON CONFLICT(school_id) DO "+
			"UPDATE SET school_name=$2,school_name_phonetic=$3,school_level_id=$4,address=$5,is_archived=$6,updated_at=now(),deleted_at=NULL", strings.Join(fields, ", "), placeHolders)
		b.Queue(query, args...)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			errors = append(errors, convertErrToErrResForEachLineCSV(err, mappers[i], "upsert"))
			continue
		}
	}
	return errors
}

func convertErrToErrResForEachLineCSV(err error, i int, method string) *ImportError {
	return &ImportError{
		RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
		Error:     fmt.Sprintf("unable to %s school_info item: %s", method, err),
	}
}

// Create creates SchoolInfo entity
func (r *SchoolInfoRepo) Create(ctx context.Context, db database.QueryExecer, e *entity.SchoolInfo) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolInfoRepo.Create")
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
		return fmt.Errorf("err create SchoolInfoRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err create SchoolInfoRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Update updates SchoolInfo entity
func (r *SchoolInfoRepo) Update(ctx context.Context, db database.QueryExecer, e *entity.SchoolInfo) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolInfoRepo.Update")
	defer span.End()

	now := time.Now()
	var err error

	err = e.UpdatedAt.Set(now)
	if err != nil {
		return err
	}

	cmdTag, err := database.UpdateFields(
		ctx,
		e,
		db.Exec,
		"school_id",
		[]string{"school_name", "school_name_phonetic", "school_level_id", "address", "is_archived", "updated_at"},
	)
	if err != nil {
		return fmt.Errorf("err update SchoolInfoRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update SchoolInfoRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *SchoolInfoRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.SchoolInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolInfoRepo.GetByIDs")
	defer span.End()

	schoolInfo := &entity.SchoolInfo{}
	fields := database.GetFieldNames(schoolInfo)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE school_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), schoolInfo.TableName())

	rows, err := db.Query(ctx, stmt, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schoolInfos := make([]*entity.SchoolInfo, 0)
	for rows.Next() {
		schoolInfo := &entity.SchoolInfo{}
		if err := rows.Scan(database.GetScanFields(schoolInfo, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		schoolInfos = append(schoolInfos, schoolInfo)
	}

	return schoolInfos, nil
}

func (r *SchoolInfoRepo) GetBySchoolPartnerIDs(ctx context.Context, db database.QueryExecer, schoolPartnerIds pgtype.TextArray) ([]*entity.SchoolInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolInfoRepo.GetByIDs")
	defer span.End()

	schoolInfo := &entity.SchoolInfo{}
	fields := database.GetFieldNames(schoolInfo)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE school_partner_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), schoolInfo.TableName())

	rows, err := db.Query(ctx, stmt, &schoolPartnerIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	schoolInfos := make([]*entity.SchoolInfo, 0)
	for rows.Next() {
		schoolInfo := &entity.SchoolInfo{}
		if err := rows.Scan(database.GetScanFields(schoolInfo, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		schoolInfos = append(schoolInfos, schoolInfo)
	}

	return schoolInfos, nil
}
