package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

// InfoNotificationRepo repo for info_notification table
type InfoNotificationRepo struct {
	InfoNotificationSQLBuilder InfoNotificationSQLBuilder
}

func (r *InfoNotificationRepo) Upsert(ctx context.Context, db database.QueryExecer, infoNotification *entities.InfoNotification) (string, error) {
	now := time.Now()
	err := multierr.Combine(
		infoNotification.CreatedAt.Set(now),
		infoNotification.UpdatedAt.Set(now),
		infoNotification.DeletedAt.Set(nil),
	)
	if err != nil {
		return "", fmt.Errorf("multierr.Combine: %w", err)
	}

	if infoNotification.NotificationID.String == "" {
		_ = infoNotification.NotificationID.Set(idutil.ULIDNow())
	}
	fields := database.GetFieldNames(infoNotification)
	values := database.GetScanFields(infoNotification, fields)
	pl := database.GeneratePlaceholders(len(fields))
	tableName := infoNotification.TableName()

	query := fmt.Sprintf(`INSERT INTO %s as noti (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__info_notifications 
		DO UPDATE SET 
			notification_msg_id = EXCLUDED.notification_msg_id,
			data = EXCLUDED.data,
			editor_id = EXCLUDED.editor_id, 
			target_groups = EXCLUDED.target_groups, 
			receiver_ids = EXCLUDED.receiver_ids,
			status = EXCLUDED.status, 
			scheduled_at = EXCLUDED.scheduled_at,
			questionnaire_id = EXCLUDED.questionnaire_id,
			updated_at = EXCLUDED.updated_at,
			is_important = EXCLUDED.is_important,
			receiver_names = EXCLUDED.receiver_names,
			generic_receiver_ids = EXCLUDED.generic_receiver_ids,
			excluded_generic_receiver_ids = EXCLUDED.excluded_generic_receiver_ids
		WHERE (noti.status = 'NOTIFICATION_STATUS_DRAFT' OR noti.status = 'NOTIFICATION_STATUS_SCHEDULED')
		AND noti.owner = EXCLUDED.owner
		AND EXCLUDED.notification_msg_id IS NOT NULL
		AND noti.deleted_at IS NULL;
		`, tableName, strings.Join(fields, ","), pl)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return "", err
	}

	if cmd.RowsAffected() == 0 {
		return "", fmt.Errorf("can not upsert notification")
	}

	return infoNotification.NotificationID.String, nil
}

type FindNotificationOption struct {
	OrderByUpdatedAt string
}

func NewFindNotificationOption() *FindNotificationOption {
	f := &FindNotificationOption{
		OrderByUpdatedAt: consts.DefaultOrder,
	}
	return f
}

type FindNotificationFilter struct {
	NotiIDs                pgtype.TextArray
	NotificationMsgIDs     pgtype.TextArray
	Status                 pgtype.TextArray
	FromScheduled          pgtype.Timestamptz
	ToScheduled            pgtype.Timestamptz
	ResourcePath           pgtype.Text
	FromSent               pgtype.Timestamptz
	ToSent                 pgtype.Timestamptz
	Type                   pgtype.Text
	Limit                  pgtype.Int8
	Offset                 pgtype.Int8
	IsLocationSelectionAll pgtype.Bool
	IsCourseSelectionAll   pgtype.Bool
	IsClassSelectionAll    pgtype.Bool
	EditorIDs              pgtype.TextArray
	Opts                   *FindNotificationOption
}

func (f *FindNotificationFilter) Validate() error {
	if f.NotiIDs.Status == pgtype.Null &&
		f.NotificationMsgIDs.Status == pgtype.Null &&
		f.Status.Status == pgtype.Null &&
		f.FromScheduled.Status == pgtype.Null &&
		f.ToScheduled.Status == pgtype.Null &&
		f.FromSent.Status == pgtype.Null &&
		f.ToSent.Status == pgtype.Null &&
		f.ResourcePath.Status == pgtype.Null &&
		f.Type.Status == pgtype.Null &&
		f.Limit.Status == pgtype.Null &&
		f.Offset.Status == pgtype.Null &&
		f.IsLocationSelectionAll.Status == pgtype.Null &&
		f.IsCourseSelectionAll.Status == pgtype.Null &&
		f.IsClassSelectionAll.Status == pgtype.Null &&
		f.EditorIDs.Status == pgtype.Null &&
		f.Opts == nil {
		return fmt.Errorf("FindNotificationFilter all field is null")
	}
	return nil
}

func NewFindNotificationFilter() *FindNotificationFilter {
	f := &FindNotificationFilter{}
	_ = f.NotiIDs.Set(nil)
	_ = f.NotificationMsgIDs.Set(nil)
	_ = f.Status.Set(nil)
	_ = f.FromScheduled.Set(nil)
	_ = f.ToScheduled.Set(nil)
	_ = f.FromSent.Set(nil)
	_ = f.ToSent.Set(nil)
	_ = f.ResourcePath.Set(nil)
	_ = f.Type.Set(nil)
	_ = f.Limit.Set(nil)
	_ = f.Offset.Set(nil)
	_ = f.IsLocationSelectionAll.Set(false)
	_ = f.IsCourseSelectionAll.Set(false)
	_ = f.IsClassSelectionAll.Set(false)
	_ = f.EditorIDs.Set(nil)
	f.Opts = NewFindNotificationOption()
	return f
}

func (r *InfoNotificationRepo) Find(ctx context.Context, db database.QueryExecer, filter *FindNotificationFilter) (entities.InfoNotifications, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	argsQuery := []interface{}{filter.NotiIDs}
	paramIndex := len(argsQuery) + 1

	filter.Opts.OrderByUpdatedAt = consts.DescendingOrder
	query := r.InfoNotificationSQLBuilder.BuildFindNotificationsByFilterSQL(filter, &paramIndex, &argsQuery)

	pagingQuery := ""
	isPaging := filter.Limit.Status == pgtype.Present && filter.Offset.Status == pgtype.Present
	if isPaging {
		paramIndexLimit := paramIndex
		paramIndexOffset := paramIndex + 1
		pagingQuery += fmt.Sprintf(`
			LIMIT $%d
			OFFSET $%d
		`, paramIndexLimit, paramIndexOffset)
		argsQuery = append(argsQuery, filter.Limit, filter.Offset)
	}

	query += pagingQuery
	notifications := entities.InfoNotifications{}
	err := database.Select(ctx, db, query, argsQuery...).ScanAll(&notifications)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *InfoNotificationRepo) CountTotalNotificationForStatus(ctx context.Context, db database.QueryExecer, filter *FindNotificationFilter) (map[string]uint32, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	argsQuery := []interface{}{filter.NotiIDs}
	paramIndex := len(argsQuery) + 1

	filter.Opts.OrderByUpdatedAt = consts.DefaultOrder
	query := r.InfoNotificationSQLBuilder.BuildFindNotificationsByFilterSQL(filter, &paramIndex, &argsQuery)

	queryCount := `
			SELECT notifications_filtered.status AS status, count(*) AS total
			FROM (
			` + query + `) AS notifications_filtered
			GROUP BY notifications_filtered.status
		`

	rows, err := db.Query(ctx, queryCount, argsQuery...)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	defer rows.Close()

	totalNotificationsForEachStatus := []*TotalNotificationForStatus{}
	totalNotificationsForAllStatus := int64(0)
	for rows.Next() {
		totalNotificationForStatus := new(TotalNotificationForStatus)
		if err := rows.Scan(&totalNotificationForStatus.Status, &totalNotificationForStatus.Total); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		totalNotificationsForEachStatus = append(totalNotificationsForEachStatus, totalNotificationForStatus)
		totalNotificationsForAllStatus += totalNotificationForStatus.Total.Int
	}

	totalNotificationsForEachStatus = append(totalNotificationsForEachStatus, &TotalNotificationForStatus{
		Status: database.Text(cpb.NotificationStatus_NOTIFICATION_STATUS_NONE.String()),
		Total:  database.Int8(totalNotificationsForAllStatus),
	})

	mapStatusAndCount := make(map[string]uint32, 0)

	for _, totalNotificationForStatus := range totalNotificationsForEachStatus {
		mapStatusAndCount[totalNotificationForStatus.Status.String] = uint32(totalNotificationForStatus.Total.Int)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return mapStatusAndCount, nil
}

func (r *InfoNotificationRepo) UpdateNotification(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text, attributes map[string]interface{}) error {
	// cover case empty
	if len(attributes) == 0 {
		return nil
	}

	counting := 1
	query := fmt.Sprintf(`UPDATE %v SET `, (&entities.InfoNotification{}).TableName())
	params := make([]interface{}, 0)
	for field, value := range attributes {
		query += fmt.Sprintf(`%v = $%v`, field, counting)
		if counting < len(attributes) {
			query += `, `
		}
		counting++
		params = append(params, value)
	}

	query += fmt.Sprintf(" WHERE notification_id = $%v", counting)
	params = append(params, notificationID)

	cmd, err := db.Exec(ctx, query, params...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (r *InfoNotificationRepo) SetStatus(ctx context.Context, db database.QueryExecer, notificationID, status pgtype.Text) error {
	e := entities.InfoNotification{}
	query := fmt.Sprintf(`UPDATE %s SET status = $1 WHERE notification_id = $2 AND deleted_at IS NULL`, e.TableName())

	cmd, err := db.Exec(ctx, query, status, notificationID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (r *InfoNotificationRepo) SetSentAt(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text) error {
	e := entities.InfoNotification{}
	query := fmt.Sprintf(`UPDATE %s SET sent_at = now() WHERE notification_id = $1 AND deleted_at IS NULL`, e.TableName())

	cmd, err := db.Exec(ctx, query, notificationID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (r *InfoNotificationRepo) DiscardNotification(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text, statues pgtype.TextArray) error {
	e := entities.InfoNotification{}
	query := fmt.Sprintf(`UPDATE %s SET status = $1, deleted_at = now() WHERE notification_id = $2 AND ($3::TEXT[] IS NULL OR status = ANY($3::TEXT[])) AND deleted_at IS NULL`, e.TableName())

	cmd, err := db.Exec(ctx, query, database.Text(cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD.String()), notificationID, statues)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (r *InfoNotificationRepo) IsNotificationDeleted(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text) (bool, error) {
	query := "SELECT status FROM info_notifications WHERE notification_id = $1 AND deleted_at IS NOT NULL"

	status := pgtype.Text{}
	err := db.QueryRow(ctx, query, notificationID).Scan(&status)
	if err != nil {
		return false, err
	}

	return status.String == cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD.String(), nil
}

type TotalNotificationForStatus struct {
	Status pgtype.Text
	Total  pgtype.Int8
}

type InfoNotificationMsgRepo struct{}

func (r *InfoNotificationMsgRepo) Upsert(ctx context.Context, db database.QueryExecer, infoNotificationMsg *entities.InfoNotificationMsg) error {
	now := time.Now()
	err := multierr.Combine(
		infoNotificationMsg.CreatedAt.Set(now),
		infoNotificationMsg.UpdatedAt.Set(now),
		infoNotificationMsg.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if infoNotificationMsg.NotificationMsgID.String == "" {
		_ = infoNotificationMsg.NotificationMsgID.Set(idutil.ULIDNow())
	}

	fields := database.GetFieldNames(infoNotificationMsg)
	values := database.GetScanFields(infoNotificationMsg, fields)
	pl := database.GeneratePlaceholders(len(fields))

	query := fmt.Sprintf(`INSERT INTO %s AS msg (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__info_notification_msgs 
		DO UPDATE SET 
			title = EXCLUDED.title, 
			content = EXCLUDED.content, 
			media_ids = EXCLUDED.media_ids, 
			updated_at = EXCLUDED.updated_at
		WHERE msg.deleted_at IS NULL;
		`, infoNotificationMsg.TableName(), strings.Join(fields, ","), pl)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not create notification")
	}

	return nil
}

func (r *InfoNotificationMsgRepo) GetByIDs(ctx context.Context, db database.QueryExecer, notiMsgIDs pgtype.TextArray) (entities.InfoNotificationMsgs, error) {
	e := &entities.InfoNotificationMsg{}
	fields := database.GetFieldNames(e)
	tableName := e.TableName()

	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE notification_msg_id = ANY($1) AND deleted_at IS NULL;
		`, strings.Join(fields, ","), tableName)

	res := entities.InfoNotificationMsgs{}
	err := database.Select(ctx, db, query, notiMsgIDs).ScanAll(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *InfoNotificationMsgRepo) GetByNotificationIDs(ctx context.Context, db database.QueryExecer, notiIDs pgtype.TextArray) (map[string]*entities.InfoNotificationMsg, error) {
	e := &entities.InfoNotificationMsg{}
	notiEnt := &entities.InfoNotification{}
	fields := database.GetFieldNames(e)
	tableName := e.TableName()

	query := fmt.Sprintf(`
		SELECT n.notification_id, nm.%s FROM %s nm 
		INNER JOIN %s n ON nm.notification_msg_id = n.notification_msg_id 
		WHERE n.notification_id = ANY($1) 
		AND n.deleted_at IS NULL
		AND nm.deleted_at IS NULL
		`, strings.Join(fields, ", nm."), tableName, notiEnt.TableName())

	res := make(map[string]*entities.InfoNotificationMsg)
	rows, err := db.Query(ctx, query, notiIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		notiID := &pgtype.Text{}
		e := &entities.InfoNotificationMsg{}
		f := []interface{}{notiID}
		f = append(f, database.GetScanFields(e, database.GetFieldNames(e))...)
		err := rows.Scan(f...)
		if err != nil {
			return nil, err
		}
		res[notiID.String] = e
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *InfoNotificationMsgRepo) GetIDsByTitle(ctx context.Context, db database.QueryExecer, title pgtype.Text) ([]string, error) {
	query := `
		SELECT notification_msg_id 
		FROM info_notification_msgs 
		WHERE title ILIKE CONCAT('%%', $1::TEXT, '%%')
			AND deleted_at IS NULL;
		`

	res := make([]string, 0)
	rows, err := db.Query(ctx, query, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		notificationMsgID := &pgtype.Text{}
		f := []interface{}{notificationMsgID}
		err := rows.Scan(f...)
		if err != nil {
			return nil, err
		}
		res = append(res, notificationMsgID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *InfoNotificationMsgRepo) SoftDelete(ctx context.Context, db database.QueryExecer, notificationMsgIDs []string) error {
	pgIDs := database.TextArray(notificationMsgIDs)

	query := `
		UPDATE info_notification_msgs AS qn
		SET deleted_at = now(), 
			updated_at = now() 
		WHERE notification_msg_id = ANY($1) 
		AND qn.deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, &pgIDs)
	if err != nil {
		return err
	}

	return nil
}

type UsersInfoNotificationRepo struct{}

func (r *UsersInfoNotificationRepo) queueUpsert(b *pgx.Batch, userInfoNotification *entities.UserInfoNotification) {
	fields := database.GetFieldNames(userInfoNotification)
	pl := database.GeneratePlaceholders(len(fields))

	query := fmt.Sprintf(`INSERT INTO %s AS user_noti (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT unique__user_id__notification_id
		DO UPDATE SET 
			status = EXCLUDED.status,
			course_ids = EXCLUDED.course_ids,
			current_grade = EXCLUDED.current_grade,
			updated_at = EXCLUDED.updated_at,
			is_individual = EXCLUDED.is_individual
		WHERE user_noti.status = 'USER_NOTIFICATION_STATUS_NEW'
		AND user_noti.deleted_at IS NULL;
		`, userInfoNotification.TableName(), strings.Join(fields, ","), pl)
	b.Queue(query, database.GetScanFields(userInfoNotification, fields)...)
}

func (r *UsersInfoNotificationRepo) Upsert(ctx context.Context, db database.QueryExecer, userInfoNotification []*entities.UserInfoNotification) error {
	b := &pgx.Batch{}
	now := time.Now()
	for _, un := range userInfoNotification {
		err := multierr.Combine(
			un.UserNotificationID.Set(idutil.ULIDNow()),
			un.CreatedAt.Set(now),
			un.UpdatedAt.Set(now),
			un.DeletedAt.Set(nil),
		)
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}
		r.queueUpsert(b, un)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		cmd, err := result.Exec()
		if err != nil || cmd.RowsAffected() != 1 {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

type FindUserNotificationFilter struct {
	UserNotificationIDs pgtype.TextArray
	UserIDs             pgtype.TextArray
	NotiIDs             pgtype.TextArray
	UserStatus          pgtype.TextArray
	Limit               pgtype.Int8
	OffsetTime          pgtype.Timestamptz
	OffsetText          pgtype.Text
	StudentID           pgtype.Text
	ParentID            pgtype.Text
	IsImportant         pgtype.Bool
}

func NewFindUserNotificationFilter() FindUserNotificationFilter {
	f := FindUserNotificationFilter{}
	_ = f.UserNotificationIDs.Set(nil)
	_ = f.UserIDs.Set(nil)
	_ = f.NotiIDs.Set(nil)
	_ = f.UserStatus.Set(nil)
	_ = f.Limit.Set(nil)
	_ = f.OffsetTime.Set(nil)
	_ = f.OffsetText.Set(nil)
	_ = f.StudentID.Set(nil)
	_ = f.ParentID.Set(nil)
	_ = f.IsImportant.Set(nil)
	return f
}

func (r *UsersInfoNotificationRepo) Find(ctx context.Context, db database.QueryExecer, filter FindUserNotificationFilter) (entities.UserInfoNotifications, error) {
	e := &entities.UserInfoNotification{}
	fields := database.GetFieldNames(e)

	res := entities.UserInfoNotifications{}
	query := fmt.Sprintf(`
		SELECT un.%s FROM users_info_notifications un 
		INNER JOIN info_notifications n ON n.notification_id = un.notification_id
		WHERE user_id = ANY($1) 
			AND ($2::TEXT[] IS NULL OR un.user_notification_id = ANY($2))
			AND ($3::TEXT[] IS NULL OR un.notification_id = ANY($3))
			AND ($4::TEXT[] IS NULL OR un.status = ANY($4)) 
			AND (($5::TIMESTAMPTZ IS NULL AND $6::TEXT IS NULL) OR (un.updated_at, un.notification_id) < ($5::TIMESTAMPTZ, $6::TEXT))
			AND (
				CASE 
					WHEN 
						$7::TEXT IS NULL AND $8::TEXT IS NULL 
					THEN 
						TRUE
					WHEN
						$7::TEXT IS NULL
					THEN 
						un.parent_id=$8
					WHEN
						$8::TEXT IS NULL
					THEN
						un.student_id=$7
					ELSE 
						un.student_id=$7 OR un.parent_id=$8
				END
			)
			AND ($9::BOOL IS NULL OR n.is_important=$9)
			AND n.deleted_at IS NULL
			AND un.deleted_at IS NULL
		ORDER BY un.updated_at DESC, un.notification_id DESC
		LIMIT $10;
		`, strings.Join(fields, ",un."))

	err := database.Select(ctx, db, query, filter.UserIDs, filter.UserNotificationIDs, filter.NotiIDs, filter.UserStatus, filter.OffsetTime, filter.OffsetText, filter.StudentID, filter.ParentID, filter.IsImportant, filter.Limit).ScanAll(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *UsersInfoNotificationRepo) SetStatus(ctx context.Context, db database.QueryExecer, userID pgtype.Text, userInfoNotificationIDs pgtype.TextArray, status pgtype.Text) error {
	e := &entities.UserInfoNotification{}
	tableName := e.TableName()

	query := fmt.Sprintf("UPDATE %s SET status = $1 WHERE user_notification_id = ANY($2) AND user_id = $3 AND deleted_at IS NULL;", tableName)

	cmd, err := db.Exec(ctx, query, status, userInfoNotificationIDs, userID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (r *UsersInfoNotificationRepo) SetStatusByNotificationIDs(ctx context.Context, db database.QueryExecer, userID pgtype.Text, notificationIDs pgtype.TextArray, status pgtype.Text) error {
	e := &entities.UserInfoNotification{}
	tableName := e.TableName()

	query := fmt.Sprintf("UPDATE %s SET status = $1 WHERE notification_id = ANY($2) AND user_id = $3 AND deleted_at IS NULL;", tableName)

	cmd, err := db.Exec(ctx, query, status, notificationIDs, userID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (r *UsersInfoNotificationRepo) CountByStatus(ctx context.Context, db database.QueryExecer, userID pgtype.Text, status pgtype.Text) (int, int, error) {
	e := &entities.UserInfoNotification{}
	tableName := e.TableName()

	query := fmt.Sprintf("SELECT COUNT(DISTINCT (user_id, notification_id,student_id)) FILTER (WHERE uin.status = $2) AS read, COUNT(DISTINCT (user_id, notification_id,student_id)) AS total FROM %s uin WHERE uin.user_id = $1 AND uin.deleted_at IS NULL", tableName)
	var st pgtype.Int8
	var total pgtype.Int8
	err := db.QueryRow(ctx, query, userID, status).Scan(&st, &total)
	if err != nil {
		return 0, 0, err
	}

	return int(st.Int), int(total.Int), nil
}

func (r *UsersInfoNotificationRepo) FindUserIDs(ctx context.Context, db database.QueryExecer, filter FindUserNotificationFilter) (map[string]entities.UserInfoNotifications, error) {
	e := &entities.UserInfoNotification{}
	fields := database.GetFieldNames(e)
	tableName := e.TableName()

	es := entities.UserInfoNotifications{}
	query := fmt.Sprintf(`
		SELECT %s FROM %s 
		WHERE notification_id = ANY($1)
		AND ($2::TEXT[] IS NULL OR status = ANY($2)) 
		AND ($3::TEXT IS NULL OR user_id > $3)
		AND deleted_at IS NULL
		ORDER BY user_id
		LIMIT $4;
		`, strings.Join(fields, ","), tableName)

	err := database.Select(ctx, db, query, filter.NotiIDs, filter.UserStatus, filter.OffsetText, filter.Limit).ScanAll(&es)
	if err != nil {
		return nil, err
	}

	res := make(map[string]entities.UserInfoNotifications)
	for _, e := range es {
		res[e.NotificationID.String] = append(res[e.NotificationID.String], e)
	}
	return res, nil
}

func (r *UsersInfoNotificationRepo) UpdateUnreadUser(ctx context.Context, db database.QueryExecer, notificationID pgtype.Text, userIDs pgtype.TextArray) error {
	e := &entities.UserInfoNotification{}
	tableName := e.TableName()

	query := fmt.Sprintf(`
		UPDATE %s SET updated_at = now() WHERE notification_id = $1 AND user_id = ANY($2) AND status = $3 AND deleted_at IS NULL;
	`, tableName)

	cmd, err := db.Exec(ctx, query, notificationID, userIDs, database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()))
	if err != nil {
		return err
	}

	if len(userIDs.Elements) > 0 && cmd.RowsAffected() == 0 {
		return fmt.Errorf("no row effected")
	}

	return nil
}

func (r *UsersInfoNotificationRepo) SetQuestionnareStatusAndSubmittedAt(ctx context.Context, db database.QueryExecer, userNotificationID string, status string, submittedAt pgtype.Timestamptz) error {
	e := &entities.UserInfoNotification{}
	tableName := e.TableName()

	query := fmt.Sprintf("UPDATE %s SET qn_status = $1, qn_submitted_at = $2 WHERE user_notification_id = $3 AND deleted_at IS NULL;", tableName)

	cmd, err := db.Exec(ctx, query, status, submittedAt, userNotificationID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func (r *UsersInfoNotificationRepo) GetNotificationIDWithFullyQnStatus(ctx context.Context, db database.QueryExecer, notificationIDs pgtype.TextArray, status pgtype.Text) ([]string, error) {
	query := `
		SELECT uin.notification_id
		FROM users_info_notifications uin 
			JOIN info_notifications inf ON uin.notification_id = inf.notification_id 
		WHERE inf.questionnaire_id IS NOT NULL
			AND inf.deleted_at IS NULL 
			AND ($1::TEXT[] IS NULL OR inf.notification_id = ANY($1::TEXT[]))
		GROUP BY uin.notification_id
		HAVING count(uin.user_notification_id) FILTER (WHERE uin.qn_status = $2) = count(uin.user_notification_id)
	`
	notificationIDsResult := make([]string, 0)
	rows, err := db.Query(ctx, query, notificationIDs, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		notificationID := &pgtype.Text{}
		f := []interface{}{notificationID}
		err := rows.Scan(f...)
		if err != nil {
			return nil, err
		}
		notificationIDsResult = append(notificationIDsResult, notificationID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notificationIDsResult, nil
}

func (r *UsersInfoNotificationRepo) SoftDeleteByNotificationID(ctx context.Context, db database.QueryExecer, notificationID string) error {
	ctx, span := interceptors.StartSpan(ctx, "UsersInfoNotificationRepo.SoftDeleteByNotificationID")
	defer span.End()

	query := `
		UPDATE users_info_notifications uin
		SET deleted_at = now()
		WHERE uin.notification_id = $1 AND uin.deleted_at IS NULL;
	`

	_, err := db.Exec(ctx, query, database.Text(notificationID))

	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}
