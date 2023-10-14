package repositories

import (
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
)

type InfoNotificationSQLBuilder struct{}

// nolint
func (repo *InfoNotificationSQLBuilder) BuildFindNotificationsByFilterSQL(filter *FindNotificationFilter, paramIndex *int, argsQuery *[]interface{}) string {
	notification := &entities.InfoNotification{}
	fields := strings.Join(database.GetFieldNames(notification), ", ifn.")

	selectQuery := fmt.Sprintf(`
		SELECT ifn.%s
		FROM info_notifications ifn 
	`, fields)

	whereQuery := `
		WHERE ($1::TEXT[] IS NULL OR ifn.notification_id = ANY($1)) 
	`

	// Assign title check condition (by notification_msg_id)
	if filter.NotificationMsgIDs.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.notification_msg_id = ANY($%d) 
		`, *paramIndex)
		*paramIndex++

		*argsQuery = append(*argsQuery, filter.NotificationMsgIDs)
	}

	// Type filter (select ALL)
	if filter.IsLocationSelectionAll.Bool {
		whereQuery += `
			AND ifn.target_groups->'location_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_ALL'
		`
	}

	if filter.IsCourseSelectionAll.Bool {
		whereQuery += `
			AND ifn.target_groups->'course_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_ALL'
		`
	}

	if filter.IsClassSelectionAll.Bool {
		whereQuery += `
			AND ifn.target_groups->'class_filter'->>'type' = 'NOTIFICATION_TARGET_GROUP_SELECT_ALL'
		`
	}

	// Assign scheduled_at condition
	// nolint
	if filter.FromScheduled.Status == pgtype.Present && filter.ToScheduled.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.scheduled_at BETWEEN $%d AND $%d
		`, *paramIndex, *paramIndex+1)
		*paramIndex += 2

		*argsQuery = append(*argsQuery, filter.FromScheduled, filter.ToScheduled)
	} else if filter.FromScheduled.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.scheduled_at >= $%d
		`, *paramIndex)
		*paramIndex++

		*argsQuery = append(*argsQuery, filter.FromScheduled)
	} else if filter.ToScheduled.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.scheduled_at <= $%d
		`, *paramIndex)
		*paramIndex++

		*argsQuery = append(*argsQuery, filter.ToScheduled)
	}

	// Assign sent_at condition
	// nolint
	if filter.FromSent.Status == pgtype.Present && filter.ToSent.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.sent_at BETWEEN $%d AND $%d
		`, *paramIndex, *paramIndex+1)
		*paramIndex += 2

		*argsQuery = append(*argsQuery, filter.FromSent, filter.ToSent)
	} else if filter.FromSent.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.sent_at >= $%d
		`, *paramIndex)
		*paramIndex++

		*argsQuery = append(*argsQuery, filter.FromSent)
	} else if filter.ToSent.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.sent_at <= $%d
		`, *paramIndex)
		*paramIndex++

		*argsQuery = append(*argsQuery, filter.ToSent)
	}

	// Assign resource_path check support for scan scheduled notification
	if filter.ResourcePath.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.resource_path = $%d 
		`, *paramIndex)
		*paramIndex++

		*argsQuery = append(*argsQuery, filter.ResourcePath)
	}

	if filter.EditorIDs.Status == pgtype.Present {
		whereQuery += fmt.Sprintf(`
			AND ifn.editor_id = ANY($%d::TEXT[]) 
		`, *paramIndex)
		*paramIndex++

		*argsQuery = append(*argsQuery, filter.EditorIDs)
	}

	whereQuery += fmt.Sprintf(`
		AND ($%d::TEXT[] IS NULL OR ifn.status = ANY($%d)) 
		AND ($%d::TEXT IS NULL OR ifn.type = $%d)
		AND ifn.deleted_at IS NULL
	`, *paramIndex, *paramIndex, *paramIndex+1, *paramIndex+1)
	*paramIndex += 2
	*argsQuery = append(*argsQuery, filter.Status, filter.Type)

	query := selectQuery + whereQuery

	if filter.Opts.OrderByUpdatedAt != consts.DefaultOrder {
		query += fmt.Sprintf(` ORDER BY ifn.updated_at %s `, filter.Opts.OrderByUpdatedAt)
	}

	return query
}
