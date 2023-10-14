package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

type TaxRepo struct{}

func (r *TaxRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, taxID string) (entities.Tax, error) {
	tax := &entities.Tax{}
	taxFieldNames, taxFieldValues := tax.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			tax_id = $1
		FOR NO KEY UPDATE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(taxFieldNames, ","),
		tax.TableName(),
	)
	row := db.QueryRow(ctx, stmt, taxID)
	err := row.Scan(taxFieldValues...)
	if err != nil {
		return entities.Tax{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *tax, nil
}

// Create creates Tax entity
func (r *TaxRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Tax) error {
	ctx, span := interceptors.StartSpan(ctx, "TaxRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.TaxID.Set(idutil.ULIDNow()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert Tax: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert Tax: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Update updates Tax entity
func (r *TaxRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.Tax) error {
	ctx, span := interceptors.StartSpan(ctx, "TaxRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "tax_id", []string{"name", "tax_percentage", "tax_category", "default_flag", "is_archived", "updated_at"})

	if err != nil {
		return fmt.Errorf("err update Tax: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Tax: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}
