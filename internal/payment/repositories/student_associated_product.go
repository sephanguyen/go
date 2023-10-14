package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type StudentAssociatedProductRepo struct{}

func (r *StudentAssociatedProductRepo) Create(ctx context.Context, db database.QueryExecer, e entities.StudentAssociatedProduct) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentAssociatedProductRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}
	_, err := database.InsertExceptOnConflictDoNothing(ctx, &e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert StudentAssociatedProduct: %w", err)
	}

	// with OnConflictDoNothingsex there will be cases where no records are added

	// if cmdTag.RowsAffected() != 1 {
	// 	return fmt.Errorf("err insert StudentAssociatedProduct: %d RowsAffected", cmdTag.RowsAffected())
	// }

	return nil
}

func (r *StudentAssociatedProductRepo) Delete(ctx context.Context, db database.QueryExecer, e entities.StudentAssociatedProduct) (err error) {
	var cmdTag pgconn.CommandTag
	ctx, span := interceptors.StartSpan(ctx, "StudentAssociatedProductRepo.Create")
	defer span.End()

	stmt := fmt.Sprintf(
		`UPDATE public.%s SET deleted_at = NOW()
		WHERE associated_product_id = $1 AND student_product_id = $2;`,
		e.TableName(),
	)
	cmdTag, err = db.Exec(ctx, stmt, e.AssociatedProductID.String, e.StudentProductID.String)
	if err != nil {
		return fmt.Errorf("err delete student_associated_product: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		err = fmt.Errorf("err update order: %d RowsAffected", cmdTag.RowsAffected())
	}
	return
}

func (r *StudentAssociatedProductRepo) GetMapAssociatedProducts(ctx context.Context, db database.QueryExecer, associatedStudentProductID string) (mapProductIDWithStudentProductIDs map[string]string, err error) {
	var (
		studentProduct entities.StudentProduct
	)
	stmt :=
		`
		SELECT sp.product_id, sp.student_product_id
		FROM 
			%s as sap
		INNER JOIN %s as sp
		ON sap.associated_product_id = sp.student_product_id 
		WHERE 
			sap.student_product_id = $1 AND sap.deleted_at IS NULL
		`
	stmt = fmt.Sprintf(
		stmt,
		(&entities.StudentAssociatedProduct{}).TableName(),
		studentProduct.TableName(),
	)
	rows, err := db.Query(ctx, stmt, associatedStudentProductID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mapProductIDWithStudentProductIDs = map[string]string{}
	for rows.Next() {
		var productID, studentProductID pgtype.Text
		err := rows.Scan(&productID, &studentProductID)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		mapProductIDWithStudentProductIDs[productID.String] = ""
	}
	return
}

func (r *StudentAssociatedProductRepo) GetAssociatedProductIDsByStudentProductID(ctx context.Context, db database.QueryExecer, entitiesID string, offset int64, limit int64) (associatedProductIDs []string, err error) {
	studentAssociatedProductEntity := &entities.StudentAssociatedProduct{}
	stmt :=
		`
		SELECT sap.*
		FROM 
			%s sap join student_product sp on sap.associated_product_id = sp.student_product_id
		WHERE 
			sap.student_product_id = $1 AND 
			sp.root_student_product_id IS NULL 
			ORDER BY sap.created_at DESC LIMIT $2 OFFSET $3
		`
	stmt = fmt.Sprintf(
		stmt,
		studentAssociatedProductEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		studentAssociatedProduct := new(entities.StudentAssociatedProduct)
		_, fieldValues := studentAssociatedProduct.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		associatedProductIDs = append(associatedProductIDs, studentAssociatedProduct.AssociatedProductID.String)
	}
	return associatedProductIDs, nil
}

func (r *StudentAssociatedProductRepo) CountAssociatedProductIDsByStudentProductID(ctx context.Context, db database.QueryExecer, entitiesID string) (total int, err error) {
	studentAssociatedProductEntity := &entities.StudentAssociatedProduct{}
	stmt :=
		`
		SELECT sap.associated_product_id
		FROM 
			%s sap join student_product sp on sap.associated_product_id = sp.student_product_id
		WHERE 
			sap.student_product_id = $1 AND 
			sp.root_student_product_id IS NULL 
		`
	stmt = fmt.Sprintf(
		stmt,
		studentAssociatedProductEntity.TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		total++
	}
	return
}
