package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"

	"github.com/jackc/pgx/v4"
)

type GradeRepo struct{}

func (g *GradeRepo) GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, pIDs []string) (og []*domain.Grade, err error) {
	ctx, span := interceptors.StartSpan(ctx, "GradeRepo.GetByPartnerInternalIDs")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE partner_internal_id = ANY($1) and deleted_at is NULL`
	d := Grade{}
	fieldNames, _ := d.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		d.TableName(),
	)
	lowerIDs := make([]string, len(pIDs))
	for i, v := range pIDs {
		lowerIDs[i] = strings.ToLower(v)
	}

	rows, err := db.Query(
		ctx,
		stmt,
		database.TextArray(lowerIDs),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.Grade
	for rows.Next() {
		var item Grade
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, item.ToGradeEntity())
	}
	return result, nil
}

func (g *GradeRepo) Import(ctx context.Context, db database.Ext, grades []*domain.Grade) error {
	ctx, span := interceptors.StartSpan(ctx, "GradeRepo.Import")
	defer span.End()
	b := &pgx.Batch{}
	for _, grade := range grades {
		grade.CreatedAt = time.Now()
		grade.UpdatedAt = time.Now()
		fields := []string{
			"grade_id",
			"name",
			"is_archived",
			"partner_internal_id",
			"sequence",
			"remarks",
			"updated_at",
			"created_at",
		}
		placeHolders := database.GeneratePlaceholders(len(fields))

		query := fmt.Sprintf("INSERT INTO grade (%s) "+
			"VALUES (%s) ON CONFLICT(grade_id) DO "+
			"UPDATE SET name = $2, is_archived = $3, partner_internal_id = $4, "+
			"sequence = $5, remarks = $6, updated_at = $7, deleted_at = NULL", strings.Join(fields, ", "), placeHolders)
		b.Queue(query, &grade.ID,
			&grade.Name,
			&grade.IsArchived,
			&grade.PartnerInternalID,
			&grade.Sequence,
			&grade.Remarks,
			&grade.UpdatedAt,
			&grade.CreatedAt)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		ct, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("grades could not be imported")
		}
	}
	return nil
}

func (g *GradeRepo) GetAll(ctx context.Context, db database.QueryExecer) (og []*domain.Grade, err error) {
	ctx, span := interceptors.StartSpan(ctx, "GradeRepo.GetAll")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE deleted_at is NULL`
	d := Grade{}
	fieldNames, _ := d.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		d.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.Grade
	for rows.Next() {
		var item Grade
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, item.ToGradeEntity())
	}
	return result, nil
}
