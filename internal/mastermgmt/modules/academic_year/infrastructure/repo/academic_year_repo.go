package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type AcademicYearRepo struct{}

func (ay *AcademicYearRepo) Insert(ctx context.Context, db database.QueryExecer, years []*domain.AcademicYear) error {
	ctx, span := interceptors.StartSpan(ctx, "AcademicYearRepo.Insert")
	defer span.End()
	b := &pgx.Batch{}
	for _, w := range years {
		academicYear, err := NewAcademicYearFromEntity(w)
		if academicYear.AcademicYearID.String == "" {
			err = multierr.Append(err, academicYear.AcademicYearID.Set(idutil.ULIDNow()))
		}
		if err != nil {
			return err
		}
		fields, args := academicYear.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			academicYear.TableName(),
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
			return fmt.Errorf("academic year is not inserted")
		}
	}
	return nil
}

func (ay *AcademicYearRepo) getAcademicYearByID(ctx context.Context, db database.Ext, id string) (*AcademicYear, error) {
	academicYear := &AcademicYear{}
	fields, values := academicYear.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM academic_year
		WHERE academic_year_id = $1
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	err := db.QueryRow(ctx, query, &id).Scan(values...)
	return academicYear, err
}

func (ay *AcademicYearRepo) GetAcademicYearByID(ctx context.Context, db database.Ext, id string) (*domain.AcademicYear, error) {
	ctx, span := interceptors.StartSpan(ctx, "AcademicYearRepo.GetAcademicYearByID")
	defer span.End()

	result, err := ay.getAcademicYearByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return result.ToAcademicYearDomain(), nil
}
