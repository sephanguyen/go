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

type AcademicWeekRepo struct{}

func (aw *AcademicWeekRepo) Insert(ctx context.Context, db database.QueryExecer, weeks []*domain.AcademicWeek) error {
	ctx, span := interceptors.StartSpan(ctx, "AcademicWeekRepo.Insert")
	defer span.End()
	b := &pgx.Batch{}
	for _, w := range weeks {
		academicWeek, err := NewAcademicWeekFromEntity(w)
		if err != nil {
			return err
		}
		fields, args := academicWeek.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT DO NOTHING`, // Don't allow to override now
			academicWeek.TableName(),
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
			return fmt.Errorf("academic week is not inserted")
		}
	}
	return nil
}

func (aw *AcademicWeekRepo) GetLocationsByAcademicWeekID(ctx context.Context, db database.QueryExecer, academicYearID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "AcademicWeekRepo.GetLocationsByAcademicWeekID")
	defer span.End()

	query := "select distinct location_id from academic_week aw where academic_year_id = $1"
	rows, err := db.Query(ctx, query, &academicYearID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	locationIds := []string{}
	for rows.Next() {
		var locationID string
		err := rows.Scan(&locationID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan location_id:%w", err)
		}
		locationIds = append(locationIds, locationID)
	}

	return locationIds, nil
}

func (aw *AcademicWeekRepo) GetAcademicWeeksByYearAndLocationIDs(ctx context.Context, db database.QueryExecer, academicYearID string, locationIDs []string) ([]*AcademicWeek, error) {
	ctx, span := interceptors.StartSpan(ctx, "AcademicWeekRepo.GetAcademicWeeksByYearAndLocationIDs")
	defer span.End()

	academicWeek := &AcademicWeek{}
	fields, _ := academicWeek.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE academic_year_id = $1 and location_id = ANY($2)
						  ORDER BY location_id, week_order`, strings.Join(fields, ","), academicWeek.TableName())

	rows, err := db.Query(ctx, query, &academicYearID, &locationIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}

	defer rows.Close()

	list := []*AcademicWeek{}
	for rows.Next() {
		item := &AcademicWeek{}
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
