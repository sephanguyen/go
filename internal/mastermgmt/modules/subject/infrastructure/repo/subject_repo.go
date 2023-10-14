package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"

	"github.com/jackc/pgx/v4"
)

type SubjectRepo struct{}

// Import all if all are valid, other case revert and import nothing.
func (s *SubjectRepo) Import(ctx context.Context, db database.Ext, subjects []*domain.Subject) error {
	ctx, span := interceptors.StartSpan(ctx, "SubjectRepo.Import")
	defer span.End()
	b := &pgx.Batch{}
	entity := &Subject{}
	for _, s := range subjects {
		s.CreatedAt = time.Now()
		s.UpdatedAt = time.Now()
		fields, _ := entity.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO %s (%s) 
			VALUES (%s) ON CONFLICT(subject_id) DO 
			UPDATE SET name = $2, updated_at = $4, deleted_at = NULL`,
			entity.TableName(), strings.Join(fields, ", "), placeHolders)
		b.Queue(query,
			&s.SubjectID,
			&s.Name,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		ct, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("subjects could not be upserted")
		}
	}
	return nil
}

func (s *SubjectRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*domain.Subject, error) {
	ctx, span := interceptors.StartSpan(ctx, "SubjectRepo.GetByIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE subject_id = ANY($1) and deleted_at is NULL`
	d := Subject{}
	fieldNames, _ := d.FieldMap()
	query = fmt.Sprintf(
		query,
		strings.Join(fieldNames, ","),
		d.TableName(),
	)
	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(ids),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.Subject
	for rows.Next() {
		var item Subject
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, item.ToEntity())
	}
	return result, nil
}

func (s *SubjectRepo) GetByNames(ctx context.Context, db database.QueryExecer, names []string) ([]*domain.Subject, error) {
	ctx, span := interceptors.StartSpan(ctx, "SubjectRepo.GetByNames")
	defer span.End()

	query := `SELECT %s FROM %s WHERE name = ANY($1) and deleted_at is NULL`
	d := Subject{}
	fieldNames, _ := d.FieldMap()
	query = fmt.Sprintf(
		query,
		strings.Join(fieldNames, ","),
		d.TableName(),
	)
	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(names),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.Subject
	for rows.Next() {
		var item Subject
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, item.ToEntity())
	}
	return result, nil
}

func (s *SubjectRepo) GetAll(ctx context.Context, db database.QueryExecer) ([]*domain.Subject, error) {
	ctx, span := interceptors.StartSpan(ctx, "SubjectRepo.GetAll")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE deleted_at is NULL`
	e := Subject{}
	fieldNames, _ := e.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		e.TableName(),
	)

	rows, err := db.Query(
		ctx,
		stmt,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.Subject
	for rows.Next() {
		var item Subject
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, item.ToEntity())
	}
	return result, nil
}
