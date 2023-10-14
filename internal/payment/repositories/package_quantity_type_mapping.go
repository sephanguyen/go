package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type PackageQuantityTypeMappingRepo struct{}

func (r *PackageQuantityTypeMappingRepo) GetByPackageTypeForUpdate(ctx context.Context, db database.QueryExecer, packageType string) (quantityType pb.QuantityType, err error) {
	packageQuantityTypeMapping := &entities.PackageQuantityTypeMapping{}
	fieldNames, fieldValues := packageQuantityTypeMapping.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			package_type = $1
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		packageQuantityTypeMapping.TableName(),
	)
	row := db.QueryRow(ctx, stmt, packageType)
	err = row.Scan(fieldValues...)
	if err != nil {
		return
	}
	quantityType = pb.QuantityType(pb.QuantityType_value[packageQuantityTypeMapping.QuantityType.String])
	return
}

func (r *PackageQuantityTypeMappingRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.PackageQuantityTypeMapping) error {
	ctx, span := interceptors.StartSpan(ctx, "PackageQuantityTypeMapping.Upsert")
	defer span.End()

	deleteQuery := `DELETE FROM package_quantity_type_mapping where package_type = $1`
	_, err := db.Exec(ctx, deleteQuery, e.PackageType)
	if err != nil {
		return fmt.Errorf("err upsert PackageQuantityTypeMappingRepo: %w", err)
	}

	insertQuery := `INSERT INTO package_quantity_type_mapping(package_type, quantity_type) VALUES ($1, $2)`
	cmdTag, err := db.Exec(ctx, insertQuery, e.PackageType, e.QuantityType)
	if err != nil {
		return fmt.Errorf("err upsert PackageQuantityTypeMappingRepo: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err upsert PackageQuantityTypeMappingRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
