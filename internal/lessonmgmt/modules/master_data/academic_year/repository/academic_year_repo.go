package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
)

type AcademicYearRepository struct{}

func (l *AcademicYearRepository) GetCurrentAcademicYear(ctx context.Context, db database.Ext) (*domain.AcademicYear, error) {
	ctx, span := interceptors.StartSpan(ctx, "AcademicYearRepository.GetCurrentAcademicYear")
	defer span.End()

	ay := &domain.AcademicYear{}
	fields, values := ay.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE now() between start_date and end_date AND deleted_at IS NULL", strings.Join(fields, ","), ay.TableName())
	err := db.QueryRow(ctx, query).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	return ay, nil
}
