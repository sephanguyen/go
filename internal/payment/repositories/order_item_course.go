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

type OrderItemCourseRepo struct{}

// MultiCreate creates OrderItems entity
func (r *OrderItemCourseRepo) MultiCreate(ctx context.Context, db database.QueryExecer, orderItemCourses []entities.OrderItemCourse) error {
	ctx, span := interceptors.StartSpan(ctx, "OrderItemCourseRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.OrderItemCourse) {
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

	for i := range orderItemCourses {
		orderItemCourse := orderItemCourses[i]
		_ = orderItemCourse.CreatedAt.Set(now)
		_ = orderItemCourse.UpdatedAt.Set(now)
		queueFn(b, &orderItemCourse)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(orderItemCourses); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("order item course not inserted")
		}
	}

	return nil
}

func (r *OrderItemCourseRepo) GetMapOrderItemCourseByOrderIDAndPackageID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	packageID string,
) (
	mapOrderItemCourse map[string]entities.OrderItemCourse,
	err error,
) {
	mapOrderItemCourse = map[string]entities.OrderItemCourse{}
	table := entities.OrderItemCourse{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = $1 and package_id = $2`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, orderID, packageID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		orderItemCourse := new(entities.OrderItemCourse)
		_, fieldValues := orderItemCourse.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = fmt.Errorf("row.Scan: %w", err)
			return
		}
		mapOrderItemCourse[orderItemCourse.CourseID.String] = *orderItemCourse
	}
	return
}

func (r *OrderItemCourseRepo) GetMapOrderItemCourseByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (
	mapOrderItemCourse map[string]entities.OrderItemCourse,
	err error,
) {
	mapOrderItemCourse = map[string]entities.OrderItemCourse{}
	table := entities.OrderItemCourse{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = $1`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, orderID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		orderItemCourse := new(entities.OrderItemCourse)
		_, fieldValues := orderItemCourse.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = fmt.Errorf("row.Scan: %w", err)
			return
		}
		mapOrderItemCourse[orderItemCourse.CourseID.String] = *orderItemCourse
	}
	return
}
