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

type ClassMemberRepo struct{}

func (rcv *ClassMemberRepo) Create(ctx context.Context, db database.QueryExecer, e *entities_bob.ClassMember) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.Create")
	defer span.End()

	now := time.Now()
	_ = e.UpdatedAt.Set(now)
	_ = e.CreatedAt.Set(now)

	cmdTag, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new ClassMember")
	}

	return nil
}

func (rcv *ClassMemberRepo) IsOwner(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userID pgtype.Text) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.IsOwner")
	defer span.End()

	e := new(entities_bob.ClassMember)
	var isOwner bool

	selectStmt := fmt.Sprintf("SELECT is_owner FROM %s WHERE class_id = $1 AND user_id = $2 AND status = '%s'", e.TableName(), entities_bob.ClassMemberStatusActive)

	row := db.QueryRow(ctx, selectStmt, &classID, &userID)

	if err := row.Scan(&isOwner); err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return false, nil
		}

		return false, errors.Wrap(err, "row.Scan")
	}

	return isOwner, nil
}

func (rcv *ClassMemberRepo) UpdateStatus(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userIDs pgtype.TextArray, status pgtype.Text) ([]*entities_bob.ClassMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.UpdateStatus")
	defer span.End()

	e := &entities_bob.ClassMember{}
	fields, _ := e.FieldMap()
	stmt := `UPDATE class_members SET status = $1, updated_at = now()
		WHERE class_id = $2 AND user_id = ANY($3)
		RETURNING ` + strings.Join(fields, ", ")
	rows, err := db.Query(ctx, stmt, &status, &classID, &userIDs)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var result []*entities_bob.ClassMember
	for rows.Next() {
		e := &entities_bob.ClassMember{}
		_, values := e.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return result, nil
}

func (rcv *ClassMemberRepo) Get(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userID, status pgtype.Text) (*entities_bob.ClassMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.Get")
	defer span.End()

	e := new(entities_bob.ClassMember)
	fields, values := e.FieldMap()

	selectStmt := fmt.Sprintf("SELECT %s "+
		"FROM %s "+
		"WHERE class_id = $1 AND user_id = $2 AND status = $3", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, selectStmt, &classID, &userID, &status).Scan(values...)
	if err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}

	return e, nil
}

func (rcv *ClassMemberRepo) FindOwner(ctx context.Context, db database.QueryExecer, classIDs pgtype.Int4Array) (mapUserIDByClass map[pgtype.Int4][]pgtype.Text, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.FindOwner")
	defer span.End()

	query := fmt.Sprintf("SELECT class_id, user_id "+
		"FROM class_members "+
		"WHERE class_id = ANY($1) AND status = '%s' AND is_owner = 'TRUE'",
		entities_bob.ClassMemberStatusActive)
	rows, err := db.Query(ctx, query, &classIDs)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	result := make(map[pgtype.Int4][]pgtype.Text)
	for rows.Next() {
		var (
			classID pgtype.Int4
			userID  pgtype.Text
		)

		if err := rows.Scan(&classID, &userID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		result[classID] = append(result[classID], userID)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return result, nil
}

func (rcv *ClassMemberRepo) Count(ctx context.Context, db database.QueryExecer, classIDs pgtype.Int4Array, userGroup pgtype.Text) (mapTotalUserByClass map[pgtype.Int4]pgtype.Int8, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.Count")
	defer span.End()

	query := fmt.Sprintf("SELECT class_id, COUNT(user_id) "+
		"FROM class_members "+
		"WHERE class_id = ANY($1) AND status = '%s' AND user_group = $2 "+
		"GROUP BY class_id",
		entities_bob.ClassMemberStatusActive)
	rows, err := db.Query(ctx, query, &classIDs, &userGroup)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	result := make(map[pgtype.Int4]pgtype.Int8)
	for rows.Next() {
		var (
			classID   pgtype.Int4
			totalUser pgtype.Int8
		)

		if err := rows.Scan(&classID, &totalUser); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		result[classID] = totalUser
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return result, nil
}

type FindClassMemberFilter struct {
	ClassIDs pgtype.Int4Array
	Status   pgtype.Text
	Group    pgtype.Text

	Limit    pgtype.Int4
	OffsetID pgtype.Text
	UserName pgtype.Text
}

const findClassMemberStmtPlt = `SELECT DISTINCT ON (concat(u.given_name || ' ', u.name) COLLATE "C", user_id) cm.%s
	FROM %s cm JOIN users u USING(user_id)
	WHERE class_id = ANY($1)
	AND (($4::text IS NULL AND $5::text IS NULL) OR (concat(u.given_name || ' ', u.name), u.user_id) > ($4::text COLLATE "C", $5::text)) 	
	AND status = $2
	AND ($3::text IS NULL OR cm.user_group = $3)
	AND cm.deleted_at IS NULL
	AND u.deleted_at IS NULL
	ORDER BY concat(u.given_name || ' ', u.name) COLLATE "C" ASC, cm.user_id ASC
	LIMIT $6::int`

func (rcv *ClassMemberRepo) Find(ctx context.Context, db database.QueryExecer, filter *FindClassMemberFilter) ([]*entities_bob.ClassMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.Find")
	defer span.End()

	e := &entities_bob.ClassMember{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf(findClassMemberStmtPlt, strings.Join(fields, ", cm."), e.TableName())
	rows, err := db.Query(ctx, query, &filter.ClassIDs, &filter.Status, &filter.Group, &filter.UserName, &filter.OffsetID, &filter.Limit)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var result []*entities_bob.ClassMember
	for rows.Next() {
		e := &entities_bob.ClassMember{}
		_, values := e.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		result = append(result, e)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return result, nil
}

func (rcv *ClassMemberRepo) FindActiveStudentMember(ctx context.Context, db database.QueryExecer, classID pgtype.Int4) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.FindStudentMember")
	defer span.End()

	query := "SELECT user_id " +
		"FROM class_members " +
		"WHERE class_id = $1 AND status = 'CLASS_MEMBER_STATUS_ACTIVE' AND user_group= 'USER_GROUP_STUDENT'"
	rows, err := db.Query(ctx, query, &classID)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		studentIDs = append(studentIDs, studentID)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return studentIDs, nil
}

func (rcv *ClassMemberRepo) FindByIDs(ctx context.Context, db database.QueryExecer, classID pgtype.Int4, userIDs pgtype.TextArray, status pgtype.Text) ([]*entities_bob.ClassMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.FindByIDs")
	defer span.End()

	e := new(entities_bob.ClassMember)
	fields := database.GetFieldNames(e)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE class_id = $1 AND user_id = ANY($2) AND status = $3", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, query, &classID, &userIDs, &status)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	members := []*entities_bob.ClassMember{}

	for rows.Next() {
		m := &entities_bob.ClassMember{}
		if err := rows.Scan(database.GetScanFields(m, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return members, nil
}

func (rcv *ClassMemberRepo) InClass(ctx context.Context, db database.QueryExecer, userID pgtype.Text, userIDs pgtype.TextArray) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.IsInClass")
	defer span.End()

	selectStmt := `
	SELECT DISTINCT user_id
	FROM class_members
	WHERE user_id = ANY($1) AND status = $2 AND class_id IN (
		SELECT class_id
		FROM class_members
		WHERE user_id = $3 AND status = $2
	)`

	var pgStatus pgtype.Text

	_ = pgStatus.Set(entities_bob.ClassMemberStatusActive)

	rows, err := db.Query(ctx, selectStmt, &userIDs, &pgStatus, &userID)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id pgtype.Text
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		ids = append(ids, id.String)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return ids, nil
}

func (rcv *ClassMemberRepo) FindUsersClass(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.FindUsersClass")
	defer span.End()

	query := `SELECT class_id FROM class_members WHERE user_id = $1 AND status='CLASS_MEMBER_STATUS_ACTIVE'`

	rows, err := db.Query(ctx, query, &userID)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	defer rows.Close()

	var classIDs []int32
	for rows.Next() {
		var classID int32
		if err := rows.Scan(&classID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		classIDs = append(classIDs, classID)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return classIDs, nil
}

func (rcv *ClassMemberRepo) ClassJoinNotIn(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, classIDs pgtype.Int4Array) ([]int32, error) {
	sql := "SELECT class_id FROM class_members WHERE user_id = $1 AND NOT(class_id = ANY($2)) AND status = 'CLASS_MEMBER_STATUS_ACTIVE'"
	rows, err := db.Query(ctx, sql, &studentID, &classIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	defer rows.Close()

	ids := []int32{}
	for rows.Next() {
		var id pgtype.Int4
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		ids = append(ids, id.Int)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", rows.Err())
	}

	return ids, nil
}

func (rcv *ClassMemberRepo) FindByClassIDsAndUserIDs(ctx context.Context, db database.QueryExecer, classIDs, userIDs pgtype.TextArray) ([]*entities_bob.ClassMemberV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.FindByClassIDsAndUserIDs")
	defer span.End()

	classMembers := &entities_bob.ClassMembersV2{}
	classMember := &entities_bob.ClassMemberV2{}
	fields, _ := classMember.FieldMap()

	stmt := fmt.Sprintf(`
	SELECT %s FROM %s
	WHERE class_id = ANY($1::_TEXT)
	AND user_id = ANY($2::_TEXT)
	AND deleted_at IS NULL`, strings.Join(fields, ", "), classMember.TableName())

	if err := database.Select(ctx, db, stmt, classIDs, userIDs).ScanAll(classMembers); err != nil {
		return nil, err
	}

	return *classMembers, nil
}

func (rcv *ClassMemberRepo) FindByUserIDsAndCourseIDs(ctx context.Context, db database.QueryExecer, userIDs, courseIDs pgtype.TextArray) ([]*entities_bob.ClassMemberV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassMemberRepo.FindByUserIDsAndCourseIDs")
	defer span.End()

	classMembers := &entities_bob.ClassMembersV2{}
	classMember := &entities_bob.ClassMemberV2{}
	fields, _ := classMember.FieldMap()

	stmt := fmt.Sprintf(`
	SELECT cm.%s FROM %s cm
	JOIN class c ON c.class_id = cm.class_id
	WHERE cm.user_id = ANY($1::_TEXT)
	AND c.course_id = ANY($2::_TEXT)
	AND cm.deleted_at IS NULL`, strings.Join(fields, ", cm."), classMember.TableName())

	if err := database.Select(ctx, db, stmt, userIDs, courseIDs).ScanAll(classMembers); err != nil {
		return nil, err
	}

	return *classMembers, nil
}
