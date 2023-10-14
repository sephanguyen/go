package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type AcademicClosedDayRepo struct{}

func (a *AcademicClosedDayRepo) Insert(ctx context.Context, db database.QueryExecer, academicClosedDays []*domain.AcademicClosedDay) error {
	ctx, span := interceptors.StartSpan(ctx, "AcademicClosedDayRepo.Insert")
	defer span.End()
	b := &pgx.Batch{}
	for _, c := range academicClosedDays {
		AcademicClosedDay, err := NewAcademicClosedDayFromEntity(c)

		if err != nil {
			return err
		}
		fields, args := AcademicClosedDay.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT DO NOTHING`, // Don't allow to override now
			AcademicClosedDay.TableName(),
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
			return fmt.Errorf("closed day is not inserted")
		}
	}
	return nil
}

func (a *AcademicClosedDayRepo) GetAcademicClosedDayByWeeks(ctx context.Context, db database.QueryExecer, weekIDs []string) ([]*AcademicClosedDay, error) {
	ctx, span := interceptors.StartSpan(ctx, "AcademicClosedDayRepo.GetAcademicClosedDayByWeeks")
	defer span.End()

	academicClosedDay := &AcademicClosedDay{}
	fields, _ := academicClosedDay.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE academic_week_id = ANY($1)", strings.Join(fields, ","), academicClosedDay.TableName())

	rows, err := db.Query(ctx, query, &weekIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	defer rows.Close()

	list := []*AcademicClosedDay{}
	for rows.Next() {
		item := &AcademicClosedDay{}
		if err = rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		list = append(list, item)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return list, nil
}
