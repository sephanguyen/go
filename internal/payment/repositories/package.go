package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type PackageRepo struct {
}

func (r *PackageRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, packageID string) (entities.Package, error) {
	pkg := &entities.Package{}
	packageFieldNames, packageFieldValues := pkg.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			package_id = $1
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(packageFieldNames, ","),
		pkg.TableName(),
	)
	row := db.QueryRow(ctx, stmt, packageID)
	err := row.Scan(packageFieldValues...)
	if err != nil {
		return entities.Package{}, err
	}
	return *pkg, nil
}

// Create creates Package entity
func (r *PackageRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Package) error {
	ctx, span := interceptors.StartSpan(ctx, "PackageRepo.Create")
	defer span.End()

	var productID pgtype.Text

	now := time.Now()
	if err := multierr.Combine(
		e.ProductID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	err := database.InsertReturningAndExcept(ctx, &e.Product, db, []string{"resource_path"}, "product_id", &productID)
	if err != nil {
		return fmt.Errorf("err insert Product: %w", err)
	}
	e.PackageID = productID

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert PackageRepo: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert PackageRepo: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

// Update updates Package entity
func (r *PackageRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.Package) error {
	ctx, span := interceptors.StartSpan(ctx, "PackageRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, &e.Product, db.Exec, "product_id", []string{
		"name",
		"tax_id",
		"available_from",
		"available_until",
		"remarks",
		"custom_billing_period",
		"billing_schedule_id",
		"disable_pro_rating_flag",
		"is_archived",
		"is_unique",
		"updated_at",
	})
	if err != nil {
		return fmt.Errorf("err update Product: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Product: %d RowsAffected", cmdTag.RowsAffected())
	}

	cmdTag, err = database.UpdateFields(ctx, e, db.Exec, "package_id", []string{
		"package_type",
		"max_slot",
		"package_start_date",
		"package_end_date",
	})
	if err != nil {
		return fmt.Errorf("err update Package: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update Package: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *PackageRepo) GetPackagesForExport(ctx context.Context, db database.QueryExecer) (packages []*entities.Package, err error) {
	packageFieldNames, _ := (&entities.Package{}).FieldMap()
	stmt := constant.GetAllQuery
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(packageFieldNames, ","),
		(&entities.Package{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		packageData := new(entities.Package)
		_, fieldValues := packageData.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		packages = append(packages, packageData)
	}
	return
}

func (r *PackageRepo) GetByID(ctx context.Context, db database.QueryExecer, packageID string) (entities.Package, error) {
	pkg := &entities.Package{}
	packageFieldNames, packageFieldValues := pkg.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			package_id = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(packageFieldNames, ","),
		pkg.TableName(),
	)
	row := db.QueryRow(ctx, stmt, packageID)
	err := row.Scan(packageFieldValues...)
	if err != nil {
		return entities.Package{}, err
	}
	return *pkg, nil
}

func (r *PackageRepo) GetByIDForUniqueProduct(ctx context.Context, db database.QueryExecer, packageID string) (entities.Package, error) {
	pkg := &entities.Package{}
	packageFieldNames, packageFieldValues := pkg.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			package_id = $1
			AND (package_type = 'PACKAGE_TYPE_ONE_TIME' OR package_type = 'PACKAGE_TYPE_SLOT_BASED')
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(packageFieldNames, ","),
		pkg.TableName(),
	)
	row := db.QueryRow(ctx, stmt, packageID)
	err := row.Scan(packageFieldValues...)
	if err != nil {
		return entities.Package{}, err
	}
	return *pkg, nil
}
