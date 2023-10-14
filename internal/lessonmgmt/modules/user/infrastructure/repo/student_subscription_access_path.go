package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgx/v4"
)

type StudentSubscriptionAccessPathRepo struct{}

func (s *StudentSubscriptionAccessPathRepo) FindLocationsByStudentSubscriptionIDs(ctx context.Context, db database.QueryExecer, studentSubscriptionIDs []string) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionAccessPathRepo.FindLocationsByStudentSubscriptionIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND student_subscription_id = ANY($1)`
	b := &StudentSubscriptionAccessPath{}
	fields, _ := b.FieldMap()

	ss := StudentSubscriptionAccessPaths{}
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fields, ", "), b.TableName()), &studentSubscriptionIDs).ScanAll(&ss)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	result := make(map[string][]string)
	for _, v := range ss {
		result[v.StudentSubscriptionID.String] = append(result[v.StudentSubscriptionID.String], v.LocationID.String)
	}

	return result, nil
}

func (s *StudentSubscriptionAccessPathRepo) FindStudentSubscriptionIDsByLocationIDs(ctx context.Context, db database.QueryExecer, locationIds []string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionAccessPathRepo.FindStudentSubscriptionIDsByLocationIDs")
	defer span.End()

	query := `SELECT %s FROM %s WHERE deleted_at IS NULL AND location_id = ANY($1)`
	b := &StudentSubscriptionAccessPath{}
	fields, _ := b.FieldMap()

	ss := StudentSubscriptionAccessPaths{}
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

func (s *StudentSubscriptionAccessPathRepo) BulkUpsertStudentSubscriptionAccessPath(ctx context.Context, db database.QueryExecer, subList domain.StudentSubscriptionAccessPaths) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionAccessPathRepo.BulkUpsertStudentSubscriptionAccessPath")
	defer span.End()

	excludedFields := []string{"deleted_at"}
	studentSubList, err := NewStudentSubscriptionAccessPathListFromDomainList(subList)
	if err != nil {
		return err
	}

	queueFn := func(b *pgx.Batch, studentSub *StudentSubscriptionAccessPath) {
		fieldNames := database.GetFieldNamesExcepts(studentSub, excludedFields)
		args := database.GetScanFields(studentSub, fieldNames)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT lesson_student_subscription_access_path_pk
			DO NOTHING`,
			studentSub.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, args...)
	}

	b := &pgx.Batch{}
	for _, studentSubAccessPath := range studentSubList {
		if err := studentSubAccessPath.PreUpsert(); err != nil {
			return fmt.Errorf("got error on PreUpsert student subscription access path: %w", err)
		}

		queueFn(b, studentSubAccessPath)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(studentSubList); i++ {
		commandTag, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("failed to bulk upsert student subscription batchResults.Exec: %w", err)
		}
		if commandTag.RowsAffected() != 1 {
			return fmt.Errorf("student subscription not inserted/updated")
		}
	}

	return nil
}

func (s *StudentSubscriptionAccessPathRepo) DeleteByStudentSubscriptionIDs(ctx context.Context, db database.QueryExecer, subIDList []string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionAccessPathRepo.DeleteByStudentSubscriptionIDs")
	defer span.End()

	query := `UPDATE lesson_student_subscription_access_path 
		SET deleted_at = NOW() 
		WHERE student_subscription_id = ANY($1) 
		AND deleted_at IS NULL`
	_, err := db.Exec(ctx, query, subIDList)

	if err != nil {
		return fmt.Errorf("error on deleting subscription access path records, db.Exec: %w", err)
	}

	return nil
}
