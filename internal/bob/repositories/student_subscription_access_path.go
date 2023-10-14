package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type StudentSubscriptionAccessPathRepo struct{}

func (s *StudentSubscriptionAccessPathRepo) Upsert(ctx context.Context, db database.Ext, cc []*entities.StudentSubscriptionAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionAccessPathRepo.Upsert")
	defer span.End()
	queue := func(b *pgx.Batch, t *entities.StudentSubscriptionAccessPath) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) 
		ON CONFLICT ON CONSTRAINT lesson_student_subscription_access_path_pk 
		DO UPDATE SET updated_at = now(), deleted_at = NULL`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		err := multierr.Combine(
			t.CreatedAt.Set(now),
			t.UpdatedAt.Set(now),
		)

		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("lesson_student_subscription_access_path_pk not inserted")
		}
	}
	return nil
}

func (s *StudentSubscriptionAccessPathRepo) DeleteByStudentSubscriptionID(ctx context.Context, db database.QueryExecer, studentSubscriptionID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionAccessPathRepo.DeleteByStudentSubscriptionID")
	defer span.End()
	sql := `UPDATE lesson_student_subscription_access_path 
	SET deleted_at = NOW() 
	WHERE student_subscription_id = $1 AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &studentSubscriptionID)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (s *StudentSubscriptionAccessPathRepo) FindStudentSubscriptionIDsByLocationIDs(ctx context.Context, db database.QueryExecer, locationIds []string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionAccessPathRepo.FindStudentSubscriptionIDsByLocationIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND location_id = ANY($1)`
	b := &entities.StudentSubscriptionAccessPath{}
	fields, _ := b.FieldMap()

	ss := entities.StudentSubscriptionAccessPaths{}
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &locationIds).ScanAll(&ss)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	result := make([]string, 0, len(ss))
	for _, v := range ss {
		result = append(result, v.StudentSubscriptionID.String)
	}
	return result, nil
}
