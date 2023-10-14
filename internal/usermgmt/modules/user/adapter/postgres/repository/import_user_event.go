package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type ImportUserEventRepo struct{}

func (r *ImportUserEventRepo) GetByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.Int8Array) ([]*entity.ImportUserEvent, error) {
	ctx, span := interceptors.StartSpan(ctx, "ImportUserEventRepo.GetByIDs")
	defer span.End()

	importUserEvent := &entity.ImportUserEvent{}

	fields := database.GetFieldNames(importUserEvent)

	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE import_user_event_id = ANY($1)`, strings.Join(fields, ","), importUserEvent.TableName())

	importUserEvents := entity.ImportUserEvents{}
	err := database.Select(ctx, db, query, ids).ScanAll(&importUserEvents)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return importUserEvents, nil
}

func (r *ImportUserEventRepo) Upsert(ctx context.Context, db database.QueryExecer, importUserEvents []*entity.ImportUserEvent) ([]*entity.ImportUserEvent, error) {
	ctx, span := interceptors.StartSpan(ctx, "ImportUserEventRepo.Upsert")
	defer span.End()

	now := time.Now()

	queueFn := func(b *pgx.Batch, u *entity.ImportUserEvent) {
		fieldsToCreate, valuesToCreate := u.FieldMap()
		if u.ID.Status == pgtype.Null {
			fieldsToCreate, valuesToCreate = database.GetFieldMapExcept(u, "import_user_event_id")
		}

		createdImportUserEvent := &entity.ImportUserEvent{}
		database.AllNullEntity(createdImportUserEvent)
		createdFields, _ := createdImportUserEvent.FieldMap()

		stmt :=
			`
		INSERT INTO 
			%s(%s)
		VALUES 
			(%s) 
		ON CONFLICT 
			ON CONSTRAINT pk__import_user_event
		DO UPDATE SET 
			updated_at = EXCLUDED.updated_at, status = EXCLUDED.status, sequence_number = EXCLUDED.sequence_number
		RETURNING
			%s;
		`

		stmt = fmt.Sprintf(
			stmt,
			u.TableName(),
			strings.Join(fieldsToCreate, ","),
			database.GeneratePlaceholders(len(fieldsToCreate)),
			strings.Join(createdFields, ","),
		)

		b.Queue(stmt, valuesToCreate...)
	}

	b := &pgx.Batch{}
	for _, importUserEvent := range importUserEvents {
		if importUserEvent.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := importUserEvent.ResourcePath.Set(resourcePath); err != nil {
				return nil, err
			}
		}

		if err := multierr.Combine(
			importUserEvent.CreatedAt.Set(now),
			importUserEvent.UpdatedAt.Set(now),
		); err != nil {
			return nil, err
		}

		queueFn(b, importUserEvent)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	importUserEventsUpserted := []*entity.ImportUserEvent{}
	for range importUserEvents {
		createdImportUserEvent := &entity.ImportUserEvent{}
		database.AllNullEntity(createdImportUserEvent)
		_, createdValues := createdImportUserEvent.FieldMap()
		err := batchResults.QueryRow().Scan(createdValues...)
		switch err {
		case nil:
			importUserEventsUpserted = append(importUserEventsUpserted, createdImportUserEvent)
		case pgx.ErrNoRows:
			return nil, errors.Wrap(err, "database.InsertReturning returns no row")
		default:
			return nil, errors.Wrap(err, "database.InsertReturning")
		}
	}

	return importUserEventsUpserted, nil
}
