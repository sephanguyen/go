package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type BillItemCourseRepo struct{}

// MultiCreate creates BillItems entity
func (r *BillItemCourseRepo) MultiCreate(ctx context.Context, db database.QueryExecer, billItemCourses []*entities.BillItemCourse, billItemSequenceNumber int32) error {
	ctx, span := interceptors.StartSpan(ctx, "BillItemCourseRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.BillItemCourse) {
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
	now := time.Now()

	for _, u := range billItemCourses {
		_ = u.CreatedAt.Set(now)
		_ = u.BillItemSequenceNumber.Set(billItemSequenceNumber)
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(billItemCourses); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("bill item course not inserted")
		}
	}

	return nil
}
