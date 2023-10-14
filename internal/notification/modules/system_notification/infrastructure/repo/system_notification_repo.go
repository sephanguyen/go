package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type SystemNotificationRepo struct{}

func (*SystemNotificationRepo) UpsertSystemNotification(ctx context.Context, db database.QueryExecer, e *model.SystemNotification) error {
	ctx, span := interceptors.StartSpan(ctx, "SystemNotificationRepo.UpsertSystemNotification")
	defer span.End()

	fieldNames := database.GetFieldNames(e)
	values := database.GetScanFields(e, fieldNames)
	pl := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`
		INSERT INTO %s as ie (%s) VALUES(%s)
		ON CONFLICT ON CONSTRAINT uk__system_notifications__reference_id
		DO UPDATE SET
			url = EXCLUDED.url,
			valid_from = EXCLUDED.valid_from,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
		WHERE ie.deleted_at IS NULL;
	`, e.TableName(), strings.Join(fieldNames, ","), pl)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed exec: %+v", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("cannot insert System Notification")
	}

	return nil
}

func (*SystemNotificationRepo) CheckUserBelongToSystemNotification(ctx context.Context, db database.QueryExecer, userID, systemNotificationID string) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "SystemNotificationRepo.FindByUserID")
	defer span.End()

	query := `
		SELECT sn.system_notification_id
		FROM system_notifications sn
		JOIN system_notification_recipients snr ON snr.system_notification_id = sn.system_notification_id
		WHERE snr.user_id = $1 AND sn.system_notification_id = $2
		AND snr.deleted_at IS NULL
		AND sn.deleted_at IS NULL;
	`

	var snID string
	err := db.QueryRow(ctx, query, database.Text(userID), database.Text(systemNotificationID)).Scan(&snID)
	if err != nil {
		return false, err
	}

	return snID == systemNotificationID, nil
}

func (*SystemNotificationRepo) SetStatus(ctx context.Context, db database.QueryExecer, systemNotificationID, status string) error {
	ctx, span := interceptors.StartSpan(ctx, "SystemNotificationRepo.SetStatus")
	defer span.End()

	query := `
		UPDATE system_notifications SET status = $1 WHERE system_notification_id = $2 AND deleted_at IS NULL;
	`

	cmd, err := db.Exec(ctx, query, database.Text(status), database.Text(systemNotificationID))
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

type FindSystemNotificationFilter struct {
	UserID    pgtype.Text
	ValidFrom pgtype.Timestamptz
	Limit     pgtype.Int8
	Offset    pgtype.Int8
	Status    pgtype.TextArray
	Language  pgtype.Text
	Keyword   pgtype.Text
}

func NewFindSystemNotificationFilter() FindSystemNotificationFilter {
	f := FindSystemNotificationFilter{}
	_ = f.UserID.Set(nil)
	_ = f.ValidFrom.Set(nil)
	_ = f.Limit.Set(nil)
	_ = f.Offset.Set(nil)
	_ = f.Status.Set(nil)
	_ = f.Language.Set(nil)
	_ = f.Keyword.Set(nil)
	return f
}

func FindSystemNotificationByFilterSQL(filter *FindSystemNotificationFilter, paramIndex *int, argsQuery *[]interface{}) string {
	systemNotification := &model.SystemNotification{}
	fields := strings.Join(database.GetFieldNames(systemNotification), ", sn.")
	selectQuery := fmt.Sprintf(`
		SELECT sn.%s
		FROM system_notifications sn
		JOIN system_notification_recipients snr ON snr.system_notification_id = sn.system_notification_id
		JOIN system_notification_contents snc on snc.system_notification_id = sn.system_notification_id
	`, fields)
	whereQuery := `
		WHERE sn.deleted_at IS NULL
		AND snr.deleted_at IS NULL
		AND snc.deleted_at IS NULL
	`
	if filter.ValidFrom.Status == pgtype.Present {
		*paramIndex++
		whereQuery += fmt.Sprintf(`
			AND sn.valid_from <= $%d
		`, *paramIndex)
		*argsQuery = append(*argsQuery, filter.ValidFrom)
	}
	if filter.UserID.Status == pgtype.Present {
		*paramIndex++
		whereQuery += fmt.Sprintf(`
			AND snr.user_id = $%d
		`, *paramIndex)
		*argsQuery = append(*argsQuery, filter.UserID)
	}
	if filter.Language.Status == pgtype.Present {
		*paramIndex++
		whereQuery += fmt.Sprintf(`
			AND snc.language = $%d
		`, *paramIndex)
		*argsQuery = append(*argsQuery, filter.Language)
	}
	if filter.Keyword.Status == pgtype.Present {
		*paramIndex++
		whereQuery += fmt.Sprintf(`
			AND snc.text ILIKE CONCAT('%%', $%d::TEXT, '%%')
		`, *paramIndex)
		*argsQuery = append(*argsQuery, filter.Keyword)
	}
	*paramIndex++
	whereQuery += fmt.Sprintf(`
		AND ($%d::TEXT[] IS NULL OR sn.status = ANY($%d))
	`, *paramIndex, *paramIndex)
	*argsQuery = append(*argsQuery, filter.Status)

	orderQuery := "ORDER BY sn.valid_from DESC"
	return selectQuery + whereQuery + orderQuery
}

func (*SystemNotificationRepo) FindSystemNotifications(ctx context.Context, db database.QueryExecer, filter *FindSystemNotificationFilter) (model.SystemNotifications, error) {
	ctx, span := interceptors.StartSpan(ctx, "SystemNotificationRepo.FindSystemNotifications")
	defer span.End()

	argsQuery := []interface{}{}
	paramIndex := len(argsQuery)
	query := FindSystemNotificationByFilterSQL(filter, &paramIndex, &argsQuery)

	pagingQuery := ""
	isPaging := filter.Limit.Status == pgtype.Present && filter.Offset.Status == pgtype.Present
	if isPaging {
		paramIndex++
		pagingQuery += fmt.Sprintf(`
			LIMIT $%d
		`, paramIndex)
		argsQuery = append(argsQuery, filter.Limit)

		paramIndex++
		pagingQuery += fmt.Sprintf(`
			OFFSET $%d
		`, paramIndex)
		argsQuery = append(argsQuery, filter.Offset)
	}
	query += pagingQuery

	systemNotifications := model.SystemNotifications{}
	err := database.Select(ctx, db, query, argsQuery...).ScanAll(&systemNotifications)
	if err != nil {
		return nil, err
	}

	return systemNotifications, nil
}

type TotalForStatus struct {
	Status pgtype.Text
	Total  pgtype.Int8
}

func (*SystemNotificationRepo) CountSystemNotifications(ctx context.Context, db database.QueryExecer, filter *FindSystemNotificationFilter) (map[string]uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "SystemNotificationRepo.CountSystemNotifications")
	defer span.End()

	_ = filter.Status.Set(nil)
	argsQuery := []interface{}{}
	paramIndex := len(argsQuery)
	query := FindSystemNotificationByFilterSQL(filter, &paramIndex, &argsQuery)

	queryCount := fmt.Sprintf(`
		SELECT filtered.status AS status, count(*) AS total
		FROM (%s) AS filtered
		GROUP BY filtered.status
	`, query)

	rows, err := db.Query(ctx, queryCount, argsQuery...)
	if err != nil {
		return nil, err
	}

	totalForAllStatus := uint32(0)
	mapStatusAndCount := make(map[string]uint32, 0)

	for rows.Next() {
		totalForStatus := new(TotalForStatus)
		if err := rows.Scan(&totalForStatus.Status, &totalForStatus.Total); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		mapStatusAndCount[totalForStatus.Status.String] = uint32(totalForStatus.Total.Int)
		totalForAllStatus += uint32(totalForStatus.Total.Int)
	}
	mapStatusAndCount[npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NONE.String()] = totalForAllStatus

	return mapStatusAndCount, nil
}

func (*SystemNotificationRepo) FindByReferenceID(ctx context.Context, db database.QueryExecer, referenceID string) (*model.SystemNotification, error) {
	ctx, span := interceptors.StartSpan(ctx, "SystemNotificationRepo.FindByReferenceID")
	defer span.End()
	e := &model.SystemNotification{}
	fields, vals := e.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s
		FROM system_notifications ie
		WHERE ie.deleted_at IS NULL AND ie.reference_id = $1;
	`, strings.Join(fields, ","))

	err := db.QueryRow(ctx, query, database.Text(referenceID)).Scan(vals...)
	if err != nil {
		// if no records found, return without error
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed QueryRow: %+v", err)
	}

	return e, nil
}
