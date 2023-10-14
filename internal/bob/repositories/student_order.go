package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentOrderRepo struct{}

func (rcv *StudentOrderRepo) Create(ctx context.Context, db database.QueryExecer, e *entities_bob.StudentOrder) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.Create")
	defer span.End()

	now := time.Now()
	_ = e.CreatedAt.Set(now)
	_ = e.UpdatedAt.Set(now)
	return database.InsertReturning(ctx, database.TrimFieldEntity{E: e, N: 1}, db, "student_order_id", &e.ID)
}

func (rcv *StudentOrderRepo) UpdateReferenceNumber(ctx context.Context, db database.QueryExecer, orderID pgtype.Int4, referenceNumber pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.Update")
	defer span.End()

	stmt := "UPDATE student_orders SET reference_number = $1, updated_at = now() " +
		"WHERE student_order_id = $2"
	cmdTag, err := db.Exec(ctx, stmt, &referenceNumber, &orderID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update reference number of student order")
	}

	return nil
}

func (rcv *StudentOrderRepo) Get(ctx context.Context, db database.QueryExecer, id int32) (*entities_bob.StudentOrder, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.Get")
	defer span.End()

	e := new(entities_bob.StudentOrder)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s "+
		"FROM student_orders "+
		"WHERE student_order_id = $1", strings.Join(fields, ","))

	row := db.QueryRow(ctx, selectStmt, &id)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

func (rcv *StudentOrderRepo) UpdateStatus(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array, currentStatus pgtype.Text, updateStatus pgtype.Text) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.UpdateStatus")
	defer span.End()

	stmt := "UPDATE student_orders SET status = $1, updated_at = now() " +
		"WHERE student_order_id = ANY($2) AND status = $3"
	cmdTag, err := db.Exec(ctx, stmt, &updateStatus, &ids, &currentStatus)
	return cmdTag.RowsAffected(), err
}

func (rcv *StudentOrderRepo) UpdatePaymentResponse(ctx context.Context, db database.QueryExecer, id pgtype.Int4, method, response, feedback pgtype.Text) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.UpdatePaymentResponse")
	defer span.End()

	stmt := "UPDATE student_orders SET gateway_response = $1, gateway_full_feedback = $2, payment_method = $3, updated_at = now() " +
		"WHERE student_order_id = $4"
	cmdTag, err := db.Exec(ctx, stmt, &response, &feedback, &method, &id)
	return cmdTag.RowsAffected(), err
}

func (rcv *StudentOrderRepo) Find(ctx context.Context, db database.QueryExecer, studentID, status pgtype.Text, packageID pgtype.Int4Array) (*entities_bob.StudentOrder, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.Find")
	defer span.End()

	e := new(entities_bob.StudentOrder)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND package_id = ANY($2) AND status = $3", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &studentID, &packageID, &status)

	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

func (rcv *StudentOrderRepo) FindByGateway(ctx context.Context, db database.QueryExecer, studentID, gateway pgtype.Text, status pgtype.Text) (*entities_bob.StudentOrder, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.FindByGateway")
	defer span.End()

	e := new(entities_bob.StudentOrder)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND gateway_name = $2 AND status = $3", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &studentID, &gateway, &status)

	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, err
	}

	return e, nil
}

func (rcv *StudentOrderRepo) Retrieve(ctx context.Context, db database.QueryExecer, studentIds pgtype.TextArray, statuses pgtype.TextArray) ([]*entities_bob.StudentOrder, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.Retrieve")
	defer span.End()

	var o entities_bob.StudentOrder
	fieldNames := database.GetFieldNames(&o)
	query := "SELECT " + strings.Join(fieldNames, ",") + " " +
		"FROM " + (&o).TableName() +
		" WHERE student_id = ANY($1) AND status = ANY($2) " +
		"ORDER BY created_at DESC"

	rows, err := db.Query(ctx, query, &studentIds, &statuses)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var orders []*entities_bob.StudentOrder
	for rows.Next() {
		order := &entities_bob.StudentOrder{}
		if err := rows.Scan(database.GetScanFields(order, fieldNames)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return orders, nil
}

func (rcv *StudentOrderRepo) ListOrderForProcessing(ctx context.Context, db database.QueryExecer, processingBefore pgtype.Timestamptz, status, gateway pgtype.Text) ([]*entities_bob.StudentOrder, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.ListPending")
	defer span.End()

	var o entities_bob.StudentOrder
	fieldNames := database.GetFieldNames(&o)
	query := "SELECT " + strings.Join(fieldNames, ",") + " " +
		"FROM " + (&o).TableName() +
		" WHERE updated_at <= $1 AND gateway_name = $2 AND status = $3 " +
		"ORDER BY created_at DESC"

	rows, err := db.Query(ctx, query, &processingBefore, &gateway, &status)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var orders []*entities_bob.StudentOrder
	for rows.Next() {
		order := &entities_bob.StudentOrder{}
		if err := rows.Scan(database.GetScanFields(order, fieldNames)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return orders, nil
}

func (rcv *StudentOrderRepo) Update(ctx context.Context, db database.QueryExecer, order *entities_bob.StudentOrder) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.Update")
	defer span.End()

	now := time.Now()
	order.UpdatedAt.Set(now)

	cmdTag, err := database.Update(ctx, order, db.Exec, "student_order_id")
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update student order")
	}

	return nil
}

func (rcv *StudentOrderRepo) CheckTransactionIDExist(ctx context.Context, db database.QueryExecer, IapTransactionID pgtype.Text) bool {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.CheckTransactionIDExist")
	defer span.End()

	var total int
	selectStmt := fmt.Sprintf("SELECT COUNT(*) " +
		"FROM student_orders " +
		"WHERE inapp_transaction_id = $1")

	_ = db.QueryRow(ctx, selectStmt, &IapTransactionID).Scan(&total)
	return total != 0
}

func (rcv *StudentOrderRepo) FindOrderByPromotionCode(ctx context.Context, db database.QueryExecer, studentID, promoCode pgtype.Text) (*entities_bob.StudentOrder, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.FindOrderByPromotionCode")
	defer span.End()

	e := new(entities_bob.StudentOrder)
	fields := database.GetFieldNames(e)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND coupon = $2 ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), e.TableName())

	row := db.QueryRow(ctx, selectStmt, &studentID, &promoCode)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNoRows
		}
		return nil, errors.Wrap(err, "row.Scan")
	}

	return e, nil
}

func (rcv *StudentOrderRepo) FindOrdersByPromotionCodeAndStatuses(ctx context.Context, db database.QueryExecer, studentID, promoCode pgtype.Text, statuses pgtype.TextArray) ([]*entities_bob.StudentOrder, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentOrderRepo.FindOrdersByPromotionCodeAndStatus")
	defer span.End()

	e := new(entities_bob.StudentOrder)
	fields := database.GetFieldNames(e)

	selectStmt := fmt.Sprintf(
		"SELECT %s "+
			"FROM %s "+
			"WHERE student_id = $1 AND coupon = $2 "+
			"AND status = ANY($3)",
		strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, selectStmt, &studentID, &promoCode, &statuses)
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryEx")
	}
	defer rows.Close()

	var orders []*entities_bob.StudentOrder
	for rows.Next() {
		order := &entities_bob.StudentOrder{}
		if err := rows.Scan(database.GetScanFields(order, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return orders, nil
}
