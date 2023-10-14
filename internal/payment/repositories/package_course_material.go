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

type PackageCourseMaterialRepo struct {
}

func (r *PackageCourseMaterialRepo) queueUpsert(b *pgx.Batch, associatedProductsByMaterial []*entities.PackageCourseMaterial) {
	queueFn := func(b *pgx.Batch, u *entities.PackageCourseMaterial) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"

		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range associatedProductsByMaterial {
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *PackageCourseMaterialRepo) Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, associatedProductsByMaterial []*entities.PackageCourseMaterial) error {
	ctx, span := interceptors.StartSpan(ctx, "PackageCourseMaterialRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(`DELETE FROM package_course_material WHERE package_id = $1;`, packageID)
	r.queueUpsert(b, associatedProductsByMaterial)

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

func (r *PackageCourseMaterialRepo) GetToTalAssociatedByCourseIDAndPackageID(ctx context.Context, db database.QueryExecer, packageID string, courseIDs []string) (total int32, err error) {
	PackageCourseMaterial := &entities.PackageCourseMaterial{}
	ProductTable := &entities.Product{}
	stmt :=
		`
		SELECT count(DISTINCT material_id)
		FROM 
			%s pcm 
		INNER JOIN %s p ON pcm.material_id = p.product_id
		WHERE 
			pcm.package_id = $1 AND
			pcm.course_id = any($2) AND
			pcm.available_from < now() AND
			pcm.available_until  > now() AND
			p.is_archived <> true AND
			p.is_unique <> true
		`

	stmt = fmt.Sprintf(
		stmt,
		PackageCourseMaterial.TableName(),
		ProductTable.TableName(),
	)
	if err := db.QueryRow(ctx, stmt, packageID, courseIDs).Scan(&total); err != nil {
		return 0, fmt.Errorf("row.Scan: %w", err)
	}
	return
}
