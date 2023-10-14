package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type PackageRepo struct {
}

func (p *PackageRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.Package) error {
	now := timeutil.Now()
	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)

	}
	fieldNames, values := e.FieldMap()
	placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15"

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (package_id)
		DO UPDATE SET
			updated_at = $15,
			country = $2,
			name = $3,
			descriptions = $4,
			price = $5,
			discounted_price = $6,
			start_at = $7,
			end_at = $8,
			duration = $9,
			prioritize_level = $10,
			properties = $11,
			is_recommended = $12,
			is_active = $13`,
		e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	ct, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return errors.New("cannot insert package")
	}
	return nil
}

func (p *PackageRepo) Get(ctx context.Context, db database.QueryExecer, ID pgtype.Text) (*entities.Package, error) {
	e := &entities.Package{}
	fields, _ := e.FieldMap()

	sql := fmt.Sprintf(`SELECT %s
			FROM %s
			WHERE package_id = $1`,
		strings.Join(fields, ","), e.TableName(),
	)

	err := database.Select(ctx, db, sql, &ID).ScanOne(e)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return e, nil
}
