package repositories

import (
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/notification/consts"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
)

type AudienceSQLBuilder struct{}

// nolint:goconst
// Will have 2 return values
// + First is a query with order statement
// + Second is a query without order statement (supporting for counting data or any query doesn't need to order to improve performance)
func (repo *AudienceSQLBuilder) BuildFindGroupAudiencesByFilterSQL(filter *FindGroupAudienceFilter, opts *FindAudienceOption, currentArgIndex *int, args *[]interface{}) (string, string) {
	studentQuerySelectClause := `SELECT s.student_id AS user_id, s.student_id AS student_id, NULL AS parent_id, s.grade_id, NULL::TEXT[] AS child_ids, 'USER_GROUP_STUDENT' AS user_group, FALSE AS is_individual`
	studentQueryFromClause := ` 
		FROM students s
	`
	studentQueryJoinClause := `
		JOIN student_enrollment_status_history sesh ON s.student_id = sesh.student_id
	`
	studentQueryWhereClause := ` 
		WHERE s.deleted_at IS NULL
			AND sesh.deleted_at IS NULL
			AND sesh.start_date <= now()
			AND (sesh.end_date >= now() OR sesh.end_date IS NULL) 
	`
	studentGroupByClause := ` 
		GROUP BY s.student_id, user_group
	`

	parentQuerySelectClause := `SELECT sp.parent_id AS user_id, sp.student_id AS student_id, sp.parent_id AS parent_id, null AS grade_id, string_to_array(sp.student_id, ',') AS child_ids, 'USER_GROUP_PARENT' AS user_group, FALSE AS is_individual`
	parentQueryFromClause := ` 
		FROM student_parents sp
	`
	parentQueryJoinClause := `
		JOIN students s ON s.student_id = sp.student_id
		JOIN student_enrollment_status_history sesh ON s.student_id = sesh.student_id
	`
	parentQueryWhereClause := ` 
		WHERE sp.deleted_at IS NULL 
			AND s.deleted_at IS NULL
			AND sesh.deleted_at IS NULL
			AND sesh.start_date <= now()
			AND (sesh.end_date >= now() OR sesh.end_date IS NULL) 
	`
	parentGroupByClause := ` 
		GROUP BY sp.parent_id, user_group, sp.student_id
	`

	if opts.OrderByName != consts.DefaultOrder || opts.IsGetName || filter.Keyword.Status == pgtype.Present {
		studentQueryJoinClause += ` JOIN users u ON u.user_id = s.student_id`
		studentGroupByClause += ` , u.name, u.email`

		parentQueryJoinClause += ` JOIN users u ON u.user_id = sp.parent_id`
		parentGroupByClause += ` , u.name, u.email`
	}

	if opts.IsGetName || filter.Keyword.Status == pgtype.Present {
		studentQuerySelectClause += `, u.name, u.email`

		parentQuerySelectClause += `, u.name, u.email`
	}

	if filter.StudentEnrollmentStatus.Status == pgtype.Present {
		studentQueryWhereClause += fmt.Sprintf(` AND sesh.enrollment_status = $%d::TEXT `, *currentArgIndex)

		parentQueryWhereClause += fmt.Sprintf(` AND sesh.enrollment_status = $%d::TEXT `, *currentArgIndex)
		*args = append(*args, filter.StudentEnrollmentStatus)
		*currentArgIndex++
	}

	if filter.Keyword.Status == pgtype.Present {
		studentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT IS NULL OR u.name ILIKE concat('%%', $%d::TEXT, '%%'))`, *currentArgIndex, *currentArgIndex)

		parentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT IS NULL OR u.name ILIKE concat('%%', $%d::TEXT, '%%'))`, *currentArgIndex, *currentArgIndex)
		*args = append(*args, filter.Keyword)
		*currentArgIndex++
	}

	if filter.CourseSelectType.Status == pgtype.Present && filter.CourseSelectType.String != cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String() {
		studentQueryJoinClause += ` JOIN notification_student_courses nsc ON nsc.student_id = s.student_id`
		studentQueryWhereClause += ` AND (nsc.start_at <= NOW() AND nsc.end_at >= NOW()) AND nsc.deleted_at IS NULL`

		parentQueryJoinClause += ` JOIN notification_student_courses nsc ON nsc.student_id = sp.student_id`
		parentQueryWhereClause += ` AND (nsc.start_at <= NOW() AND nsc.end_at >= NOW()) AND nsc.deleted_at IS NULL`
		if filter.CourseIDs.Status == pgtype.Present {
			studentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR nsc.course_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
			parentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR nsc.course_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
			*args = append(*args, filter.CourseIDs)
			*currentArgIndex++
		}
	}

	if filter.ClassSelectType.Status == pgtype.Present && filter.ClassSelectType.String != cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String() {
		studentQueryJoinClause += ` JOIN notification_class_members ncm ON ncm.student_id = s.student_id`
		studentQueryWhereClause += ` 
			AND ((
					ncm.start_at <= NOW() 
					OR ncm.start_at IS NULL
				)
				AND (
					ncm.end_at >= NOW() 
					OR ncm.end_at IS NULL 
					OR ncm.end_at = '0001-01-01 00:00:00.000 +0000'
				)
			) 
			AND ncm.deleted_at IS NULL
		`

		parentQueryJoinClause += ` JOIN notification_class_members ncm ON ncm.student_id = sp.student_id`
		parentQueryWhereClause += ` 
			AND ((
					ncm.start_at <= NOW() 
					OR ncm.start_at IS NULL
				)
				AND (
					ncm.end_at >= NOW() 
					OR ncm.end_at IS NULL 
					OR ncm.end_at = '0001-01-01 00:00:00.000 +0000'
				)
			) 
			AND ncm.deleted_at IS NULL
		`

		if filter.ClassIDs.Status == pgtype.Present {
			studentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR ncm.class_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
			parentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR ncm.class_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
			*args = append(*args, filter.ClassIDs)
			*currentArgIndex++
		}
	}

	// if location empty -> should be get 0 records
	if filter.LocationIDs.Status != pgtype.Undefined {
		studentQueryWhereClause += fmt.Sprintf(` 
			AND ($%d::TEXT[] IS NULL OR sesh.location_id = ANY($%d::TEXT[]))
		`, *currentArgIndex, *currentArgIndex)

		parentQueryWhereClause += fmt.Sprintf(` 
			AND ($%d::TEXT[] IS NULL OR sesh.location_id = ANY($%d::TEXT[]))
		`, *currentArgIndex, *currentArgIndex)

		if filter.CourseSelectType.String != cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String() {
			studentQueryWhereClause += fmt.Sprintf(` 
				AND ($%d::TEXT[] IS NULL OR nsc.location_id = ANY($%d::TEXT[]))
			`, *currentArgIndex, *currentArgIndex)

			parentQueryWhereClause += fmt.Sprintf(` 
				AND ($%d::TEXT[] IS NULL OR nsc.location_id = ANY($%d::TEXT[]))
			`, *currentArgIndex, *currentArgIndex)
		}

		if filter.ClassSelectType.String != cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String() {
			studentQueryWhereClause += fmt.Sprintf(` 
				AND ($%d::TEXT[] IS NULL OR ncm.location_id = ANY($%d::TEXT[]))
			`, *currentArgIndex, *currentArgIndex)

			parentQueryWhereClause += fmt.Sprintf(` 
				AND ($%d::TEXT[] IS NULL OR ncm.location_id = ANY($%d::TEXT[]))
			`, *currentArgIndex, *currentArgIndex)
		}

		*args = append(*args, filter.LocationIDs)
		*currentArgIndex++
	}

	if filter.GradeSelectType.String == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String() {
		studentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR s.grade_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
		parentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR s.grade_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
		*args = append(*args, filter.GradeIDs)
		*currentArgIndex++
	}

	if filter.IncludeUserIds.Status == pgtype.Present {
		studentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR s.student_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
		parentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR sp.parent_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
		*args = append(*args, filter.IncludeUserIds)
		*currentArgIndex++
	}

	if filter.ExcludeUserIds.Status == pgtype.Present {
		studentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR s.student_id != ALL($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
		parentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR sp.parent_id != ALL($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
		*args = append(*args, filter.ExcludeUserIds)
		*currentArgIndex++
	}

	if filter.SchoolSelectType.Status == pgtype.Present && filter.SchoolSelectType.String != consts.TargetGroupSelectTypeNone.String() {
		studentQueryJoinClause += ` JOIN school_history sh ON s.student_id = sh.student_id`
		studentQueryWhereClause += ` AND (sh.is_current IS TRUE)`

		parentQueryJoinClause += ` JOIN school_history sh ON s.student_id = sh.student_id`
		parentQueryWhereClause += ` AND (sh.is_current IS TRUE)`

		if filter.SchoolIDs.Status == pgtype.Present {
			studentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR sh.school_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
			parentQueryWhereClause += fmt.Sprintf(` AND ($%d::TEXT[] IS NULL OR sh.school_id = ANY($%d::TEXT[]))`, *currentArgIndex, *currentArgIndex)
			*args = append(*args, filter.SchoolIDs)
			*currentArgIndex++
		}
	}

	studentQuery := studentQuerySelectClause + studentQueryFromClause + studentQueryJoinClause + studentQueryWhereClause + studentGroupByClause
	parentQuery := parentQuerySelectClause + parentQueryFromClause + parentQueryJoinClause + parentQueryWhereClause + parentGroupByClause

	var queryStatements []string
	for _, userGroup := range filter.UserGroups.Elements {
		switch userGroup.String {
		case cpb.UserGroup_USER_GROUP_STUDENT.String():
			queryStatements = append(queryStatements, `(`+studentQuery+`)`)
		case cpb.UserGroup_USER_GROUP_PARENT.String():
			queryStatements = append(queryStatements, `(`+parentQuery+`)`)
		}
	}
	query := strings.Join(queryStatements, ` UNION `)

	queryWithoutOrder := query

	if query != "" && opts.OrderByName != consts.DefaultOrder {
		query += fmt.Sprintf(` ORDER BY name %s `, opts.OrderByName)
	}

	if query == "" {
		*currentArgIndex = 1
		*args = []interface{}{}
	}

	return query, queryWithoutOrder
}

// nolint:goconst
func (repo *AudienceSQLBuilder) BuildFindIndividualAudiencesByFilterSQL(filter *FindIndividualAudienceFilter, opts *FindAudienceOption, currentArgIndex *int, args *[]interface{}) string {
	studentQuerySelectClause := `SELECT s.student_id AS user_id, s.student_id AS student_id, NULL AS parent_id, s.grade_id, NULL::TEXT[] AS child_ids, 'USER_GROUP_STUDENT' AS user_group, TRUE AS is_individual
	`
	studentQueryFromClause := ` 
		FROM students s
	`
	studentQueryJoinClause := `
		JOIN user_access_paths uap ON uap.user_id = s.student_id
	`
	studentQueryWhereClause := ` 
		WHERE s.deleted_at IS NULL
			AND uap.deleted_at IS NULL
	`
	studentGroupByClause := ` 
		GROUP BY s.student_id, user_group
	`

	parentQuerySelectClause := `SELECT p.parent_id AS user_id, NULL as student_id, p.parent_id AS parent_id, NULL AS grade_id, NULL::TEXT[] AS child_ids, 'USER_GROUP_PARENT' as user_group, TRUE AS is_individual`
	parentQueryFromClause := ` 
		FROM parents p
	`
	parentQueryJoinClause := `
		JOIN user_access_paths uap ON uap.user_id = p.parent_id
	`
	parentQueryWhereClause := ` 
		WHERE p.deleted_at IS NULL
			AND uap.deleted_at IS NULL
	`
	parentGroupByClause := ` 
		GROUP BY p.parent_id, user_group
	`

	if opts.OrderByName != consts.DefaultOrder || opts.IsGetName {
		studentQueryJoinClause += ` JOIN users u ON u.user_id = s.student_id`
		studentGroupByClause += ` , u.name, u.email`

		parentQueryJoinClause += ` JOIN users u ON u.user_id = p.parent_id`
		parentGroupByClause += ` , u.name, u.email`
	}

	if opts.IsGetName {
		studentQuerySelectClause += `, u.name, u.email`

		parentQuerySelectClause += `, u.name, u.email`
	}

	if filter.EnrollmentStatuses.Status == pgtype.Present {
		studentQueryJoinClause += ` JOIN student_enrollment_status_history sesh ON s.student_id = sesh.student_id`
		studentQueryWhereClause += fmt.Sprintf(` AND sesh.deleted_at IS NULL
			AND sesh.enrollment_status = ANY($%d::TEXT[])
			AND sesh.start_date <= now()
			AND (sesh.end_date >= now() OR sesh.end_date IS NULL)
		`, *currentArgIndex)

		*args = append(*args, filter.EnrollmentStatuses)
		*currentArgIndex++
	}

	if filter.UserIDs.Status != pgtype.Undefined {
		studentQueryWhereClause += fmt.Sprintf(`
			AND s.student_id = ANY($%d::TEXT[])
		`, *currentArgIndex)

		parentQueryWhereClause += fmt.Sprintf(`
			AND p.parent_id = ANY($%d::TEXT[])
		`, *currentArgIndex)

		*args = append(*args, filter.UserIDs)
		*currentArgIndex++
	}

	if filter.LocationIDs.Status != pgtype.Undefined {
		studentQueryWhereClause += fmt.Sprintf(`
			AND uap.location_id = ANY($%d::TEXT[])
		`, *currentArgIndex)

		if filter.EnrollmentStatuses.Status == pgtype.Present {
			studentQueryWhereClause += fmt.Sprintf(` AND sesh.location_id = ANY($%d::TEXT[])
			`, *currentArgIndex)
		}

		parentQueryWhereClause += fmt.Sprintf(`
			AND uap.location_id = ANY($%d::TEXT[])
		`, *currentArgIndex)

		*args = append(*args, filter.LocationIDs)
		*currentArgIndex++
	}

	studentQuery := studentQuerySelectClause + studentQueryFromClause + studentQueryJoinClause + studentQueryWhereClause + studentGroupByClause
	parentQuery := parentQuerySelectClause + parentQueryFromClause + parentQueryJoinClause + parentQueryWhereClause + parentGroupByClause
	query := `(
			` + studentQuery + `
		) UNION (
			` + parentQuery + `
		)
	`

	return query
}
