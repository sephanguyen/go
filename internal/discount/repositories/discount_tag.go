package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type DiscountTagRepo struct {
}

func (r *DiscountTagRepo) GetByID(
	ctx context.Context,
	db database.QueryExecer,
	discountTagID string,
) (
	*entities.DiscountTag,
	error,
) {
	discountTag := &entities.DiscountTag{}
	discountTagFieldNames, discountFieldValues := discountTag.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE
			discount_tag_id = $1
		AND 
			is_archived = FALSE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(discountTagFieldNames, ","),
		discountTag.TableName(),
	)
	row := db.QueryRow(ctx, stmt, discountTagID)
	err := row.Scan(discountFieldValues...)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return discountTag, nil
}

func (r *DiscountTagRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.DiscountTag) error {
	ctx, span := interceptors.StartSpan(ctx, "DiscountTag.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.DiscountTagID.Set(idutil.ULIDNow()), e.CreatedAt.Set(now), e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}
	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert DiscountTag: %v", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert DiscountTag: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

func (r *DiscountTagRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.DiscountTag) error {
	ctx, span := interceptors.StartSpan(ctx, "DiscountTag.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "discount_tag_id", []string{"discount_tag_name", "selectable", "updated_at", "is_archived"})
	if err != nil {
		return fmt.Errorf("err update DiscountTag: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update DiscountTag: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}
