package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type BillItemAccountCategoryRepo struct{}

func (r *BillItemAccountCategoryRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, billItemAccountCategories []*entities.BillItemAccountCategory) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "BillItemAccountCategoryRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.BillItemAccountCategory) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[:len(fields)-1] // excepts resource_path field
		valuesExceptResourcePath := values[:len(values)-1]

		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			u.TableName(),
			strings.Join(fieldsExceptResourcePath, ","),
			placeHolders,
		)

		b.Queue(stmt, valuesExceptResourcePath...)
	}

	b := &pgx.Batch{}

	for _, u := range billItemAccountCategories {
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(billItemAccountCategories); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("bill item account categories not inserted")
		}
	}

	return nil
}
