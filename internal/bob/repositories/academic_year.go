package repositories

import (
	"context"
	"fmt"
	"strings"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
)

type AcademicYearRepo struct{}

func (r *AcademicYearRepo) Create(ctx context.Context, db database.QueryExecer, a *entities_bob.AcademicYear) error {
	ctx, span := interceptors.StartSpan(ctx, "AcademicYearRepo.Create")
	defer span.End()

	now := timeutil.Now()
	_ = a.CreatedAt.Set(now)
	_ = a.UpdatedAt.Set(now)

	if _, err := database.InsertIgnoreConflict(ctx, a, db.Exec); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

func (r *AcademicYearRepo) Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities_bob.AcademicYear, error) {
	ctx, span := interceptors.StartSpan(ctx, "AcademicYearRepo.Get")
	defer span.End()

	e := &entities_bob.AcademicYear{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE academic_year_id = $1", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}
