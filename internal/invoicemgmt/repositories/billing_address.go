package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type BillingAddressRepo struct {
}

func (r *BillingAddressRepo) FindByUserID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BillingAddress, error) {
	ctx, span := interceptors.StartSpan(ctx, "BillingAddressRepo.FindByID")
	defer span.End()

	e := &entities.BillingAddress{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, studentID).ScanOne(e)

	switch err {
	case nil:
		return e, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByUserID BillingAddressRepo: %w", err)
	}
}

func (r *BillingAddressRepo) FindByID(ctx context.Context, db database.QueryExecer, billingAddressID string) (*entities.BillingAddress, error) {
	ctx, span := interceptors.StartSpan(ctx, "BillingAddressRepo.FindByID")
	defer span.End()

	billingAddress := &entities.BillingAddress{}
	fields, _ := billingAddress.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE billing_address_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), billingAddress.TableName())

	err := database.Select(ctx, db, query, billingAddressID).ScanOne(billingAddress)

	switch err {
	case nil:
		return billingAddress, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByID BillingAddressRepo: %w", err)
	}
}

func (r *BillingAddressRepo) Upsert(ctx context.Context, db database.QueryExecer, billingAddresses ...*entities.BillingAddress) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingAddressRepo.Upsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, billingAddress *entities.BillingAddress) {
		fields := database.GetFieldNames(billingAddress)
		fields = utils.RemoveStrFromSlice(fields, "resource_path")
		values := database.GetScanFields(billingAddress, fields)

		placeHolders := database.GeneratePlaceholders(len(fields))

		stmt :=
			`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT billing_address__pk 
			DO UPDATE SET 
				postal_code = EXCLUDED.postal_code, 
				prefecture_code = EXCLUDED.prefecture_code, 
				city = EXCLUDED.city, 
				street1 = EXCLUDED.street1, 
				street2 = EXCLUDED.street2, 
				updated_at = now(), 
				deleted_at = NULL
			`

		stmt = fmt.Sprintf(stmt, billingAddress.TableName(), strings.Join(fields, ","), placeHolders)
		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	for _, billingAddress := range billingAddresses {
		queueFn(batch, billingAddress)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(billingAddresses); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when upserting billing_address")
		}
	}

	return nil
}

func (r *BillingAddressRepo) SoftDelete(ctx context.Context, db database.QueryExecer, billingAddressIDs ...string) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingAddressRepo.SoftDelete")
	defer span.End()

	stmt :=
		`
		UPDATE %s SET deleted_at = now() WHERE billing_address_id = ANY($1) AND deleted_at IS NULL
		`

	stmt = fmt.Sprintf(stmt, (&entities.BillingAddress{}).TableName())
	_, err := db.Exec(ctx, stmt, database.TextArray(billingAddressIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}
