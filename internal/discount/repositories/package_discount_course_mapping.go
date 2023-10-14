package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type PackageDiscountCourseMappingRepo struct {
}

func (r *PackageDiscountCourseMappingRepo) queueUpsert(b *pgx.Batch, packageDiscountCourseMappings []*entities.PackageDiscountCourseMapping) {
	queueFn := func(b *pgx.Batch, u *entities.PackageDiscountCourseMapping) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"
		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range packageDiscountCourseMappings {
		_ = u.CreatedAt.Set(now)
		_ = u.UpdatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *PackageDiscountCourseMappingRepo) Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, e []*entities.PackageDiscountCourseMapping) error {
	ctx, span := interceptors.StartSpan(ctx, "PackageDiscountCourseMappingRepo.Upsert")
	defer span.End()

	// Create a pgx.Batch to queue multiple queries together.
	b := &pgx.Batch{}

	b.Queue(`DELETE FROM package_discount_course_mapping WHERE package_id = $1;`, packageID)

	r.queueUpsert(b, e)

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}
