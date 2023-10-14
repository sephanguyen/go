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

type NotificationDateRepo struct{}

func (r *NotificationDateRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.NotificationDate) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "NotificationDateRepo.Upsert")
	defer span.End()

	deleteQuery := `DELETE FROM notification_date where order_type = $1`
	_, err = db.Exec(ctx, deleteQuery, e.OrderType.String)
	if err != nil {
		return fmt.Errorf("err delete NotificationDateRepo: %w", err)
	}

	now := time.Now()
	_ = multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if e.NotificationDateID.Status != pgtype.Present || e.NotificationDateID.String == "" {
		_ = multierr.Combine(
			e.NotificationDateID.Set(idutil.ULIDNow()),
		)
	}
	var fieldNames []string
	updateCommand := "order_type = $2, notification_date = $3, is_archived = $4 ,updated_at = $6"
	fieldNames = database.GetFieldNamesExcepts(e, []string{"resource_path"})
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) 
		VALUES (%s) ON CONFLICT
		ON CONSTRAINT notification_date__notification_date_id__pk DO UPDATE
		SET %s`, e.TableName(), strings.Join(fieldNames, ","), placeHolders, updateCommand)
	args := database.GetScanFields(e, fieldNames)
	commandTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error when upsert notification date: %v", err)
	}

	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("upsert notification date have no row affected")
	}
	return
}

func (r *NotificationDateRepo) GetByOrderType(
	ctx context.Context,
	db database.QueryExecer,
	orderType string,
) (notificationDate entities.NotificationDate, err error) {
	notificationDate = entities.NotificationDate{}
	fieldNames, fieldValues := notificationDate.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s
				FROM "%s" 
				WHERE order_type = $1
				`,
		strings.Join(fieldNames, ","),
		notificationDate.TableName(),
	)
	row := db.QueryRow(ctx, stmt, orderType)

	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
	}
	return
}
