package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type PackageCourseFeeRepo struct {
}

func (r *PackageCourseFeeRepo) queueUpsert(b *pgx.Batch, associatedProductsByFee []*entities.PackageCourseFee) {
	queueFn := func(b *pgx.Batch, u *entities.PackageCourseFee) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"

		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range associatedProductsByFee {
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *PackageCourseFeeRepo) Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, associatedProductsByFee []*entities.PackageCourseFee) error {
	ctx, span := interceptors.StartSpan(ctx, "PackageCourseFeeRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(`DELETE FROM package_course_fee WHERE package_id = $1;`, packageID)
	r.queueUpsert(b, associatedProductsByFee)

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

func (r *PackageCourseFeeRepo) GetToTalAssociatedByCourseIDAndPackageID(ctx context.Context, db database.QueryExecer, packageID string, courseIDs []string) (total int32, err error) {
	packageCourseFee := &entities.PackageCourseFee{}
	ProductTable := &entities.Product{}
	stmt :=
		`
		SELECT count(DISTINCT fee_id)
		FROM 
			%s pcf
		INNER JOIN %s p ON pcf.fee_id = p.product_id
		WHERE 
			pcf.package_id = $1 AND
			pcf.course_id = any($2) AND
			pcf.available_from < now() AND
			pcf.available_until  > now() AND
			p.is_archived <> true AND
			p.is_unique <> true
		`

	stmt = fmt.Sprintf(
		stmt,
		packageCourseFee.TableName(),
		ProductTable.TableName(),
	)
	if err := db.QueryRow(ctx, stmt, packageID, courseIDs).Scan(&total); err != nil {
		return 0, fmt.Errorf("row.Scan: %w", err)
	}

	return
}
