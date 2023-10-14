package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type OrderRepo struct{}

type OrderListFilter struct {
	StudentName          string
	OrderStatus          string
	OrderTypes           []string
	OrderIDs             []string
	LocationIDs          []string
	CreatedFrom          time.Time
	CreatedTo            time.Time
	IsReviewed           *bool
	IsStudentNotEnrolled bool
	Limit                *int64
	Offset               *int64
}

func (r *OrderRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Order) error {
	ctx, span := interceptors.StartSpan(ctx, "OrderRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	fieldNames, values := e.FieldMap()
	fieldNamesNeedInsert := fieldNames[0 : len(fieldNames)-2] // excepts resource_path and  order_sequence_number field
	valuesNeedInsert := values[0 : len(fieldNames)-2]         // excepts resource_path and  order_sequence_number field
	placeHolders := database.GeneratePlaceholders(len(fieldNamesNeedInsert))

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s ;", "public.order", strings.Join(fieldNamesNeedInsert, ","), placeHolders, "order_sequence_number")

	row := db.QueryRow(ctx, query, valuesNeedInsert...)
	var orderSequenceNumber pgtype.Int4
	err := row.Scan(&orderSequenceNumber)
	if err != nil {
		return err
	}
	err = e.OrderSequenceNumber.Set(orderSequenceNumber)
	if err != nil {
		return fmt.Errorf("OrderSequenceNumber.Set: %w", err)
	}
	return err
}

func (r *OrderRepo) UpdateIsReviewFlagByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	isReviewFlag bool,
	orderVersionNumber int32,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "OrderRepo.UpdateIsReviewFlagByOrderID")
	defer span.End()

	e := &entities.Order{}

	stmt := fmt.Sprintf(
		`UPDATE public.%s SET is_reviewed = $1
		WHERE order_id = $2 AND
		version_number = $3;`,
		e.TableName(),
	)
	cmdTag, err := db.Exec(ctx, stmt, isReviewFlag, orderID, orderVersionNumber)
	if err != nil {
		return fmt.Errorf("err update order: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		err = fmt.Errorf("err update order or out of version: %d RowsAffected", cmdTag.RowsAffected())
	}

	return
}

func (r *OrderRepo) UpdateOrderStatusByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	orderStatus string,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "OrderRepo.UpdateOrderStatusByOrderID")
	defer span.End()

	e := &entities.Order{}
	stmt := fmt.Sprintf(`
		UPDATE public.%s SET order_status = '%s', updated_at = now()
		WHERE order_id = '%s';
	`, e.TableName(), orderStatus, orderID)
	cmdTag, err := db.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("err update order: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		err = fmt.Errorf("err update order: %d RowsAffected", cmdTag.RowsAffected())
	}

	return
}

func (r *OrderRepo) UpdateOrderStatusByOrderIDAndVersion(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	orderStatus string,
	orderVersionNumber int32,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "OrderRepo.UpdateOrderStatusByOrderID")
	defer span.End()

	e := &entities.Order{}
	stmt := fmt.Sprintf(`
		UPDATE public.%s SET order_status = '%s', updated_at = now()
		WHERE order_id = '%s' AND version_number = $1;
	`, e.TableName(), orderStatus, orderID)
	cmdTag, err := db.Exec(ctx, stmt, orderVersionNumber)
	if err != nil {
		return fmt.Errorf("err update order: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		err = fmt.Errorf("err update order or out version number: %d RowsAffected", cmdTag.RowsAffected())
	}

	return
}

func (r *OrderRepo) GetOrderByIDForUpdate(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (order entities.Order, err error) {
	fieldNames, fieldValues := order.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s 
				FROM "%s" 
				WHERE order_id = $1
				FOR NO KEY UPDATE
				`,
		strings.Join(fieldNames, ","),
		order.TableName(),
	)
	row := db.QueryRow(ctx, stmt, orderID)
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}

func (r *OrderRepo) GetOrderTypeByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (orderType string, err error) {
	order := entities.Order{}
	stmt := fmt.Sprintf(
		`SELECT order_type
				FROM "%s" 
				WHERE order_id = $1
				`,
		order.TableName(),
	)
	row := db.QueryRow(ctx, stmt, orderID)
	err = row.Scan(&orderType)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}

func (r *OrderRepo) GetOrderByStudentIDAndLocationIDsPaging(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationIDs []string,
	from int64,
	limit int64,
) (
	orders []*entities.Order,
	err error,
) {
	var rows pgx.Rows
	table := entities.Order{}
	fieldNames, _ := table.FieldMap()

	if len(locationIDs) == 0 {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE student_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
			`,
			strings.Join(fieldNames, ","),
			table.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, limit, from)
		if err != nil {
			return nil, err
		}
	} else {
		stmt := fmt.Sprintf(
			`SELECT %s FROM "%s" WHERE student_id = $1 AND location_id = ANY($2)
			ORDER BY created_at DESC LIMIT $3 OFFSET $4
			`,
			strings.Join(fieldNames, ","),
			table.TableName(),
		)

		rows, err = db.Query(ctx, stmt, studentID, locationIDs, limit, from)
		if err != nil {
			return nil, err
		}
	}

	defer rows.Close()

	orders = make([]*entities.Order, 0, limit)
	for rows.Next() {
		order := new(entities.Order)
		_, fieldValues := order.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		orders = append(orders, order)
	}
	return
}

func (r *OrderRepo) CountOrderByStudentIDAndLocationIDs(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationIDs []string,
) (
	total int,
	err error,
) {
	table := entities.Order{}
	var rows pgx.Rows
	if len(locationIDs) == 0 {
		stmt := fmt.Sprintf(
			`SELECT order_sequence_number FROM "%s" WHERE student_id = $1`,
			table.TableName(),
		)
		rows, err = db.Query(ctx, stmt, studentID)
		if err != nil {
			return
		}
	} else {
		stmt := fmt.Sprintf(
			`SELECT order_sequence_number FROM "%s" WHERE student_id = $1 AND location_id = ANY($2)`,
			table.TableName(),
		)
		rows, err = db.Query(ctx, stmt, studentID, locationIDs)
		if err != nil {
			return
		}
	}

	defer rows.Close()

	for rows.Next() {
		total++
	}
	return
}

func (r *OrderRepo) GetAll(ctx context.Context, db database.QueryExecer) ([]*entities.Order, error) {
	table := entities.Order{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s"`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*entities.Order
	for rows.Next() {
		order := new(entities.Order)
		_, fieldValues := order.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, order)
	}
	return result, nil
}

func (r *OrderRepo) GetOrderStatsByFilter(ctx context.Context, db database.QueryExecer, filter OrderListFilter) (orderStats entities.OrderStats, err error) {
	getListOfOrdersQuery, args := r.buildGetListOfOrdersWithFilterQuery(filter)
	_, fieldValues := (&orderStats).FieldOrderStatsMap()
	stmt := fmt.Sprintf(
		`
		SELECT  
			COUNT(filtered_order.order_id) AS total_items,
			SUM(CASE         	     	WHEN filtered_order.order_status  = 'ORDER_STATUS_SUBMITTED' THEN 1 ELSE 0
        	   END) AS total_of_submitted,
        	SUM(CASE 
        	     	WHEN filtered_order.order_status  = 'ORDER_STATUS_PENDING' THEN 1 ELSE 0
        	   END) AS total_of_pending,
        	SUM(CASE 
        	    	WHEN filtered_order.order_status  = 'ORDER_STATUS_REJECTED' THEN 1 ELSE 0
        	   END) AS total_of_rejected,
        	SUM(CASE 
        	     	WHEN filtered_order.order_status  = 'ORDER_STATUS_VOIDED' THEN 1 ELSE 0
        	   END) AS total_of_voided,
        	SUM(CASE 
        	     	WHEN filtered_order.order_status  = 'ORDER_STATUS_INVOICED' THEN 1 ELSE 0
        	   END) AS total_of_invoiced,
        	SUM(CASE 
        	    	WHEN filtered_order.is_reviewed  = false AND filtered_order.order_status = 'ORDER_STATUS_SUBMITTED' THEN 1 ELSE 0
        	   END) AS total_of_order_need_to_review
		FROM (%s) AS filtered_order; 
				`,
		getListOfOrdersQuery,
	)
	row := db.QueryRow(ctx, stmt, args...)
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf("row.Scan: %w", err)
		return
	}
	return
}

func (r *OrderRepo) GetOrdersByFilter(ctx context.Context, db database.QueryExecer, filter OrderListFilter) (orders []entities.Order, err error) {
	getListOfOrdersQuery, args := r.buildGetListOfOrdersWithFilterQuery(filter)

	rows, err := db.Query(ctx, getListOfOrdersQuery, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		order := new(entities.Order)
		_, fieldValues := order.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		orders = append(orders, *order)
	}
	return
}

func (r *OrderRepo) buildGetListOfOrdersWithFilterQuery(filter OrderListFilter) (query string, args []interface{}) {
	argsIndex := 1
	table := entities.Order{}
	fieldNames, _ := table.FieldMap()

	studentHistoryIDsByEnrollmentStatusJoinQuery := ""
	if filter.IsStudentNotEnrolled {
		studentHistoryIDsByEnrollmentStatusQuery := `SELECT sesh.student_id FROM student_enrollment_status_history sesh 
WHERE sesh.student_id NOT IN(
	SELECT DISTINCT ON  (sesh.student_id, sesh.location_id)
	sesh.student_id 
	FROM student_enrollment_status_history sesh
	WHERE 
		(sesh.enrollment_status IN (
		'STUDENT_ENROLLMENT_STATUS_ENROLLED',
		'STUDENT_ENROLLMENT_STATUS_LOA')
		AND now() >= sesh.start_date)
	OR 
		(sesh.enrollment_status IN (
		'STUDENT_ENROLLMENT_STATUS_WITHDRAWN',
		'STUDENT_ENROLLMENT_STATUS_GRADUATED',
		'STUDENT_ENROLLMENT_STATUS_LOA')
		AND now() < sesh.start_date)
	ORDER BY sesh.student_id, sesh.location_id, sesh.start_date DESC
) 
GROUP BY sesh.student_id`
		studentHistoryIDsByEnrollmentStatusJoinQuery = fmt.Sprintf(`JOIN (%s) as shbes ON shbes.student_id = o.student_id`, studentHistoryIDsByEnrollmentStatusQuery)
	}

	fieldNamesWithPrefix := sliceutils.Map(fieldNames, func(fieldName string) string {
		return fmt.Sprintf("o.%s", fieldName)
	})

	query = fmt.Sprintf(`
			SELECT %s FROM "%s" o %s
			WHERE student_full_name ~* '.*%s.*'`, strings.Join(fieldNamesWithPrefix, ","), table.TableName(), studentHistoryIDsByEnrollmentStatusJoinQuery, filter.StudentName)

	if filter.OrderStatus != "" {
		query += fmt.Sprintf(" AND o.order_status = $%d", argsIndex)
		argsIndex += 1
		args = append(args, filter.OrderStatus)
	}
	if len(filter.OrderTypes) > 0 {
		args = append(args, filter.OrderTypes)
		query += fmt.Sprintf(" AND o.order_type = ANY($%d)", argsIndex)
		argsIndex += 1
	}
	if len(filter.OrderIDs) > 0 {
		query += fmt.Sprintf(" AND o.order_id = ANY($%d)", argsIndex)
		argsIndex += 1
		args = append(args, filter.OrderIDs)
	}
	if len(filter.LocationIDs) > 0 {
		query += fmt.Sprintf(" AND o.location_id = ANY($%d)", argsIndex)
		argsIndex += 1
		args = append(args, filter.LocationIDs)
	}
	if filter.IsReviewed != nil {
		if *filter.IsReviewed {
			query += fmt.Sprintf(" AND o.is_reviewed = true")
		} else {
			query += fmt.Sprintf(" AND o.is_reviewed = false")
		}
	}
	if !filter.CreatedFrom.IsZero() {
		query += fmt.Sprintf(" AND o.created_at >= $%d", argsIndex)
		argsIndex += 1
		args = append(args, filter.CreatedFrom)
	}
	if !filter.CreatedTo.IsZero() {
		query += fmt.Sprintf(" AND o.created_at <= $%d", argsIndex)
		argsIndex += 1
		args = append(args, filter.CreatedTo)
	}

	query += fmt.Sprintf(" ORDER BY o.created_at DESC")
	if filter.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argsIndex)
		argsIndex += 1
		args = append(args, *filter.Offset)
	}
	if filter.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argsIndex)
		argsIndex += 1
		args = append(args, *filter.Limit)
	}
	return
}

func (r *OrderRepo) GetOrderByStudentIDAndLocationIDForResume(
	ctx context.Context,
	db database.QueryExecer,
	studentID string,
	locationID string,
) (orderID string, err error) {
	order := entities.Order{}
	stmt := fmt.Sprintf(
		`SELECT order_id
				FROM "%s" 
				WHERE student_id = $1 AND
				location_id = $2 AND
				order_type = $3
				ORDER BY created_at DESC LIMIT 1
				`,
		order.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, locationID, pb.OrderType_ORDER_TYPE_LOA.String())
	err = row.Scan(&orderID)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}

func (r *OrderRepo) GetLatestOrderByStudentIDAndLocationIDAndOrderType(
	ctx context.Context,
	db database.QueryExecer,
	studentID, locationID, orderType string,
) (order entities.Order, err error) {
	fieldNames, fieldValues := order.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s
				FROM "%s" 
				WHERE student_id = $1 AND
				location_id = $2 AND
				order_type = $3
				ORDER BY created_at DESC LIMIT 1
				`,
		strings.Join(fieldNames, ","),
		order.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, locationID, orderType)
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}

func (r *OrderRepo) GetLatestOrderByStudentIDAndLocationID(
	ctx context.Context,
	db database.QueryExecer,
	studentID, locationID string,
) (order entities.Order, err error) {
	fieldNames, fieldValues := order.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s
				FROM "%s" 
				WHERE student_id = $1 AND
				location_id = $2
				ORDER BY created_at DESC LIMIT 1
				`,
		strings.Join(fieldNames, ","),
		order.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID, locationID)
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}
