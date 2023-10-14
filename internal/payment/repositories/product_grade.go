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

type ProductGradeRepo struct{}

func (r *ProductGradeRepo) GetByGradeAndProductIDForUpdate(ctx context.Context, db database.QueryExecer, gradeID string, productID string) (entities.ProductGrade, error) {
	productGrade := &entities.ProductGrade{}
	productGradeFieldNames, productGradeFieldValues := productGrade.FieldMap()
	stmt := `
		SELECT %s
		FROM 
			%s
		WHERE 
			product_id = $1 and grade_id = $2
		FOR NO KEY UPDATE`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productGradeFieldNames, ","),
		productGrade.TableName(),
	)
	row := db.QueryRow(ctx, stmt, productID, gradeID)
	err := row.Scan(productGradeFieldValues...)
	if err != nil {
		return entities.ProductGrade{}, err
	}
	return *productGrade, nil
}

func (r *ProductGradeRepo) queueUpsert(b *pgx.Batch, productGrades []*entities.ProductGrade) {
	queueFn := func(b *pgx.Batch, u *entities.ProductGrade) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"

		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range productGrades {
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *ProductGradeRepo) Upsert(ctx context.Context, db database.QueryExecer, productID pgtype.Text, productGrades []*entities.ProductGrade) error {
	ctx, span := interceptors.StartSpan(ctx, "ProductGradeRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(`DELETE FROM product_grade WHERE product_id = $1;`, productID)
	r.queueUpsert(b, productGrades)

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

func (r *ProductGradeRepo) GetGradeIDsByProductID(ctx context.Context, db database.QueryExecer, productID string) (gradeIDs []string, err error) {
	productGrade := &entities.ProductGrade{}
	productGradeFieldNames, productGradeFieldValues := productGrade.FieldMap()
	stmt := `
		SELECT %s
		FROM %s
		WHERE
		    product_id = $1`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(productGradeFieldNames, ","),
		productGrade.TableName(),
	)
	rows, err := db.Query(ctx, stmt, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(productGradeFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		gradeIDs = append(gradeIDs, productGrade.GradeID.String)
	}
	return
}
