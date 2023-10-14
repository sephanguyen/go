package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ClassMemberRepo struct{}

func (c *ClassMemberRepo) UpsertClassMembers(ctx context.Context, db database.QueryExecer, classMembers []*domain.ClassMember) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.UpsertClassMembers")
	defer span.End()
	b := &pgx.Batch{}
	for _, c := range classMembers {
		classMember, err := NewClassMemberFromEntity(c)
		if err != nil {
			return err
		}
		fields, args := classMember.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__class_member DO UPDATE
		SET updated_at = $5, deleted_at = $6`, classMember.TableName(), strings.Join(fields, ","), placeHolders)
		b.Queue(query, args...)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("class member is not upserted")
		}
	}
	return nil
}

func (c *ClassMemberRepo) GetByClassIDAndUserIDs(ctx context.Context, db database.QueryExecer, classID string, userIDs []string) (map[string]*domain.ClassMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.GetByClassIDAndUserIDs")
	defer span.End()
	cm := &ClassMember{}
	fields := database.GetFieldNames(cm)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE class_id = $1 AND user_id = ANY ($2) AND deleted_at IS NULL", strings.Join(fields, ","), cm.TableName())
	rows, err := db.Query(ctx, query, &classID, &userIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	result := make(map[string]*domain.ClassMember)
	for rows.Next() {
		cm := new(ClassMember)
		if err := rows.Scan(database.GetScanFields(cm, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		result[cm.UserID.String] = cm.ToClassMemberEntity()
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return result, nil
}

func (c *ClassMemberRepo) GetByClassIDAndUserID(ctx context.Context, db database.QueryExecer, classID, userID string) (*domain.ClassMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.GetByClassIDAndUserID")
	defer span.End()
	classMember := &ClassMember{}
	fields, values := classMember.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE class_id = $1 and user_id = $2 AND deleted_at is null",
		strings.Join(fields, ", "), classMember.TableName())
	err := db.QueryRow(ctx, query, &classID, &userID).Scan(values...)
	if err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return classMember.ToClassMemberEntity(), nil
}

func (c *ClassMemberRepo) UpsertClassMember(ctx context.Context, db database.QueryExecer, classMember *domain.ClassMember) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.UpsertClassMember")
	defer span.End()

	dto, err := NewClassMemberFromEntity(classMember)
	if err != nil {
		return err
	}
	if dto.ClassMemberID.Status != pgtype.Present {
		dto.ClassMemberID = database.Text(idutil.ULIDNow())
	}
	fields, args := dto.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__class_member DO UPDATE
		SET updated_at = $5, deleted_at = $6, start_date = $7, end_date = $8`, dto.TableName(), strings.Join(fields, ","), placeHolders)
	cmdTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("class member is not upserted")
	}
	return nil
}

func (c *ClassMemberRepo) DeleteByUserIDAndClassID(ctx context.Context, db database.QueryExecer, userID, classID string) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.DeleteByUserIDAndClassID")
	defer span.End()
	query := fmt.Sprintf("UPDATE class_member SET deleted_at = now() WHERE user_id = $1 AND class_id = $2 AND deleted_at IS NULL")
	_, err := db.Exec(ctx, query, userID, classID)
	return err
}

func (c *ClassMemberRepo) FindStudentIDWithCourseIDsByClassIDs(ctx context.Context, db database.QueryExecer, classIds []string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.FindStudentIDWithCourseIDsByClassIDs")
	defer span.End()

	query := `SELECT lm.user_id, l.course_id FROM class_member lm join class l on l.class_id = lm.class_id WHERE lm.class_id = ANY($1) and lm.deleted_at is null and l.deleted_at is null`
	rows, err := db.Query(ctx, query, &classIds)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	defer rows.Close()
	var studentIDWithCourses []string
	for rows.Next() {
		var studentID, courseID string
		if err := rows.Scan(&studentID, &courseID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		studentIDWithCourses = append(studentIDWithCourses, studentID, courseID)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return studentIDWithCourses, nil
}

func (c *ClassMemberRepo) GetByUserAndCourse(ctx context.Context, db database.QueryExecer, userID, courseID string) (map[string]*domain.ClassMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.GetByUserAndCourse")
	defer span.End()
	cm := &ClassMember{}
	fields := database.GetFieldNames(cm)
	rows, err := db.Query(ctx, fmt.Sprintf(`SELECT cm.%s FROM class_member cm JOIN class c ON c.class_id = cm.class_id
	where cm.user_id = $1 and c.course_id = $2 and c.deleted_at is null and cm.deleted_at is null`, strings.Join(fields, ", cm.")), &userID, &courseID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	results := make(map[string]*domain.ClassMember)
	for rows.Next() {
		cm := new(ClassMember)
		if err := rows.Scan(database.GetScanFields(cm, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		results[cm.UserID.String] = cm.ToClassMemberEntity()
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return results, nil
}

const findClassMemberStmtPlt = `SELECT DISTINCT ON (concat(u.given_name || ' ', u.name) COLLATE "C", user_id) cm.%s
	FROM %s cm JOIN users u USING(user_id)
	WHERE class_id = ANY($1)
	AND (($2::text IS NULL AND $3::text IS NULL) OR (concat(u.given_name || ' ', u.name), u.user_id) > ($2::text COLLATE "C", $3::text)) 	
	AND cm.deleted_at IS NULL
	AND u.deleted_at IS NULL
	ORDER BY concat(u.given_name || ' ', u.name) COLLATE "C" ASC, cm.user_id ASC
	LIMIT $4::int`

func (c *ClassMemberRepo) RetrieveByClassIDs(ctx context.Context, db database.QueryExecer, filter *queries.FindClassMemberFilter) (classMembers []*domain.ClassMember, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.Find")
	defer span.End()

	cm := &ClassMember{}
	fields, _ := cm.FieldMap()
	query := fmt.Sprintf(findClassMemberStmtPlt, strings.Join(fields, ", cm."), cm.TableName())
	rows, err := db.Query(ctx, query, &filter.ClassIDs, &filter.UserName, &filter.OffsetID, &filter.Limit)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()
	var cms []*domain.ClassMember
	for rows.Next() {
		cm := new(ClassMember)
		if err := rows.Scan(database.GetScanFields(cm, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		cms = append(cms, cm.ToClassMemberEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return cms, nil
}

// nolint
func (c *ClassMemberRepo) RetrieveByClassMembers(ctx context.Context, db database.QueryExecer, filter *queries.RetrieveByClassMembersFilter) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.RetrieveByClassMembers")
	defer span.End()

	if filter.Limit == 0 {
		filter.Limit = 100
	}

	var (
		args []interface{}
		stmt string
	)

	switch {
	case len(filter.ClassIDs.Elements) > 0 && len(filter.StudentTagIDs.Elements) > 0:
		stmt = `SELECT DISTINCT cm.user_id
		FROM class_member cm
		JOIN tagged_user tu ON cm.user_id = tu.user_id
		JOIN unnest(%s::_TEXT) user_id_from_course ON user_id_from_course = cm.user_id
		WHERE cm.class_id = ANY(%s::_TEXT)
		AND tu.tag_id = ANY(%s::_TEXT)
		AND cm.deleted_at IS NULL
		AND tu.deleted_at IS NULL `
		args = append(args, &filter.StudentIDs, &filter.ClassIDs, &filter.StudentTagIDs)

		switch {
		case filter.Unassigned:
			stmt += `AND cm.user_id NOT IN (
				SELECT student_id 
				FROM school_history sh
				JOIN school_info si ON sh.school_id = si.school_id 
				AND sh.deleted_at IS NULL
				AND sh.is_current = TRUE
				AND si.deleted_at IS NULL
			) `

		case filter.SchoolID.Status == pgtype.Present:
			stmt += `AND cm.user_id IN (
				SELECT student_id 
				FROM school_history sh
				JOIN school_info si on sh.school_id = si.school_id 
				AND sh.deleted_at is null
				AND sh.is_current = TRUE
				AND si.deleted_at is null
				AND si.school_id = %s::TEXT
			)`
			args = append(args, &filter.SchoolID)

		default:
			// no-op
		}

		stmt += ` AND (%s::TEXT IS NULL OR cm.user_id > %s::TEXT)
		ORDER BY cm.user_id
		LIMIT %d`
		args = append(args, &filter.Offset, &filter.Offset)

	case len(filter.ClassIDs.Elements) == 0 && len(filter.StudentTagIDs.Elements) == 0:
		switch {
		case filter.Unassigned:
			stmt = `SELECT student_id
			FROM unnest(%s::_TEXT) student_id
			WHERE student_id NOT IN (
				SELECT student_id 
				FROM school_history sh
				JOIN school_info si ON sh.school_id = si.school_id 
				AND sh.deleted_at IS NULL
				AND sh.is_current = TRUE
				AND si.deleted_at IS NULL
			)`
			args = append(args, &filter.StudentIDs)

		case filter.SchoolID.Status == pgtype.Present:
			stmt = `SELECT student_id 
			FROM school_history sh
			JOIN unnest(%s::_TEXT) user_id_from_course ON user_id_from_course = sh.student_id
			JOIN school_info si ON sh.school_id = si.school_id 
			AND sh.deleted_at IS NULL
			AND sh.is_current = TRUE
			AND si.deleted_at IS NULL
			AND si.school_id = %s::TEXT`
			args = append(args, &filter.StudentIDs, &filter.SchoolID)

		default:
			stmt = `SELECT student_id
			FROM unnest(%s::_TEXT) student_id
			WHERE 1=1`
			args = append(args, &filter.StudentIDs)
		}

		stmt += ` AND (%s::TEXT IS NULL OR student_id > %s::TEXT)
		ORDER BY student_id
		LIMIT %d`
		args = append(args, &filter.Offset, &filter.Offset)

	default:
		if len(filter.ClassIDs.Elements) > 0 {
			stmt = `SELECT DISTINCT user_id
		FROM class_member
		JOIN unnest(%s::_TEXT) user_id_from_course ON user_id_from_course = user_id
		WHERE class_id = ANY(%s::_TEXT) AND deleted_at IS NULL `
			args = append(args, &filter.StudentIDs, &filter.ClassIDs)
		}

		if len(filter.StudentTagIDs.Elements) > 0 {
			stmt = `SELECT DISTINCT user_id
		FROM tagged_user
		JOIN unnest(%s::_TEXT) user_id_from_course ON user_id_from_course = user_id
		WHERE tag_id = ANY(%s::_TEXT)
		AND deleted_at IS NULL `
			args = append(args, &filter.StudentIDs, &filter.StudentTagIDs)
		}

		switch {
		case filter.Unassigned:
			stmt += `AND user_id NOT IN (
				SELECT student_id 
				FROM school_history sh
				JOIN school_info si ON sh.school_id = si.school_id 
				AND sh.deleted_at IS NULL
				AND sh.is_current = TRUE
				AND si.deleted_at IS NULL
			) `

		case filter.SchoolID.Status == pgtype.Present:
			stmt += `AND user_id IN (
				SELECT student_id 
				FROM school_history sh
				JOIN school_info si on sh.school_id = si.school_id 
				AND sh.deleted_at is null
				AND sh.is_current = TRUE
				AND si.deleted_at is null
				AND si.school_id = %s::TEXT
			)`
			args = append(args, &filter.SchoolID)

		default:
			// no-op
		}

		stmt += ` AND (%s::TEXT IS NULL OR user_id > %s::TEXT)
		ORDER BY user_id
		LIMIT %d`
		args = append(args, &filter.Offset, &filter.Offset)
	}

	placeHolders := make([]interface{}, 0, len(args))
	for i := 0; i < len(args); i++ {
		placeHolders = append(placeHolders, fmt.Sprintf("$%d", i+1))
	}
	placeHolders = append(placeHolders, filter.Limit)
	query := fmt.Sprintf(stmt, placeHolders...)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}

	defer rows.Close()
	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return userIDs, nil
}
