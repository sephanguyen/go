package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LessonMemberRepo struct{}

func (r *LessonMemberRepo) UpsertQueue(b *pgx.Batch, e *entities.LessonMember) {
	fields, values := e.FieldMap()

	placeHolders := database.GeneratePlaceholders(len(fields))
	sql := fmt.Sprintf("INSERT INTO %s (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT pk__lesson_members DO "+
		"UPDATE SET updated_at = $3, deleted_at = NULL", e.TableName(), strings.Join(fields, ", "), placeHolders)

	b.Queue(sql, values...)
}

func (r *LessonMemberRepo) UpdateFieldsQueue(b *pgx.Batch, e *entities.LessonMember, updateFields entities.UpdateLessonMemberFields) {
	placeHolders := database.GenerateUpdatePlaceholders(updateFields.StringArray(), 3)
	stmt := fmt.Sprintf("UPDATE %s SET updated_at = now(), %s WHERE lesson_id = $1 and user_id = $2 and deleted_at is NULL", e.TableName(), placeHolders)

	args := database.GetScanFields(e, updateFields.StringArray())
	args = append([]interface{}{e.LessonID, e.UserID}, args...)
	b.Queue(stmt, args...)
}

func (r *LessonMemberRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, lessonIDs pgtype.TextArray) error {
	sql := `UPDATE lesson_members
		SET deleted_at = NOW()
		WHERE user_id = $1 AND lesson_id = ANY($2) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &studentID, &lessonIDs)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (r *LessonMemberRepo) Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.LessonMember, error) {
	e := &entities.LessonMember{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf(`SELECT %s
		FROM %s
		WHERE user_id = $1 AND deleted_at IS NULL`, strings.Join(fields, ","), e.TableName())

	members := entities.LessonMembers{}
	err := database.Select(ctx, db, query, &studentID).ScanAll(&members)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return members, nil
}

func (r *LessonMemberRepo) CourseAccessible(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]string, error) {
	sql := `SELECT l.course_id
		FROM lesson_members m JOIN lessons l
			ON m.lesson_id = l.lesson_id
				AND m.user_id = $1
				AND m.deleted_at IS NULL
				AND l.deleted_at IS NULL`

	lessons := entities.Lessons{}
	err := database.Select(ctx, db, sql, &studentID).ScanAll(&lessons)
	if err != nil {
		return nil, fmt.Errorf("db.Select: %w", err)
	}

	results := make([]string, 0, len(lessons))
	for _, l := range lessons {
		results = append(results, l.CourseID.String)
	}

	return results, nil
}

type ListStudentsByLessonArgs struct {
	LessonID pgtype.Text
	Limit    uint32

	// used for pagination
	UserName pgtype.Text
	UserID   pgtype.Text
}

func (r *LessonMemberRepo) ListStudentsByLessonID(ctx context.Context, db database.QueryExecer, args *ListStudentsByLessonArgs) ([]*entities.User, error) {
	fields, _ := (&entities.User{}).FieldMap()

	sql := fmt.Sprintf(
		`SELECT u.%s
        FROM lesson_members AS lm
        INNER JOIN users AS u ON lm.user_id = u.user_id
        WHERE lm.lesson_id = $1
        AND u.user_group = 'USER_GROUP_STUDENT'
        AND (($2::text IS NULL AND $3::text IS NULL) OR (concat(u.given_name || ' ', u.name), u.user_id) > ($2::text COLLATE "C", $3::text)) 
        AND lm.deleted_at IS NULL
        ORDER BY concat(u.given_name || ' ', u.name) COLLATE "C" ASC, u.user_id ASC
        LIMIT $4`, strings.Join(fields, ", u."))

	users := entities.Users{}
	err := database.Select(ctx, db, sql, &args.LessonID, &args.UserName, &args.UserID, &args.Limit).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("db.Select: %w", err)
	}

	return users, nil
}

func (r *LessonMemberRepo) GetLessonMemberStatesByUser(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) (entities.LessonMemberStates, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMemberStatesByUser")
	defer span.End()

	fields, _ := (&entities.LessonMemberState{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_members_states
		WHERE lesson_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	states := entities.LessonMemberStates{}
	err := database.Select(ctx, db, query, &lessonID, &userID).ScanAll(&states)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return states, nil
}

type MemberStatesFilter struct {
	LessonID  pgtype.Text
	UserID    pgtype.Text
	StateType pgtype.Text
}

func (r *LessonMemberRepo) GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *MemberStatesFilter) (entities.LessonMemberStates, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMemberStatesWithParams")
	defer span.End()
	fields, _ := (&entities.LessonMemberState{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s FROM lesson_members_states WHERE deleted_at IS NULL`, strings.Join(fields, ","))
	args := []interface{}{}
	if filter.LessonID.Get() != nil {
		query += fmt.Sprintf(" AND lesson_id = $%d ", len(args)+1)
		args = append(args, &filter.LessonID)
	}
	if filter.UserID.Get() != nil {
		query += fmt.Sprintf(" AND user_id = $%d ", len(args)+1)
		args = append(args, &filter.UserID)
	}
	if filter.StateType.Get() != nil {
		query += fmt.Sprintf(" AND state_type = $%d ", len(args)+1)
		args = append(args, &filter.StateType)
	}
	states := entities.LessonMemberStates{}
	err := database.Select(ctx, db, query, args...).ScanAll(&states)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return states, nil
}

func (r *LessonMemberRepo) GetLessonMemberStates(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.LessonMemberStates, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMemberStates")
	defer span.End()

	fields, _ := (&entities.LessonMemberState{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_members_states
		WHERE lesson_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	states := entities.LessonMemberStates{}
	err := database.Select(ctx, db, query, &lessonID).ScanAll(&states)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return states, nil
}

func (r *LessonMemberRepo) UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpsertAllLessonMemberStateByStateType")
	defer span.End()

	fields, values := state.FieldMap()
	placeHolders := database.GeneratePlaceholdersWithFirstIndex(3, len(fields)) // 2 first placeHolders is lesson_id and state_type
	query := fmt.Sprintf(
		"INSERT INTO lesson_members_states (lesson_id, user_id, state_type, updated_at, %s) "+
			"SELECT lesson_id , user_id , $2, now(), %s FROM lesson_members "+
			"WHERE lesson_id = $1 "+
			"ON CONFLICT ON CONSTRAINT lesson_members_states_pk DO "+
			"UPDATE SET updated_at = now(), %s WHERE lesson_members_states.state_type = $2 AND lesson_members_states.deleted_at IS NULL",
		strings.Join(fields, ", "), placeHolders, database.GenerateUpdatePlaceholders(fields, 3),
	)

	args := []interface{}{lessonID, stateType}
	args = append(args, values...)
	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *LessonMemberRepo) UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpsertLessonMemberState")
	defer span.End()

	fields, args := state.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	query := fmt.Sprintf("INSERT INTO lesson_members_states (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT lesson_members_states_pk DO "+
		"UPDATE SET updated_at = now(), deleted_at = null, bool_value = $7, string_array_value = $8", strings.Join(fields, ", "), placeHolders)

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *LessonMemberRepo) UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpsertAllLessonMemberStateByStateType")
	defer span.End()
	fields, values := state.FieldMap()
	placeHolders := database.GeneratePlaceholdersWithFirstIndex(4, len(fields)) // 3 first placeHolders is lesson_id, state_type and userIds
	query := fmt.Sprintf(
		"INSERT INTO lesson_members_states (lesson_id, user_id, state_type, updated_at, %s) "+
			"SELECT lesson_id , user_id , $2, now(), %s FROM lesson_members "+
			"WHERE lesson_id = $1 and user_id = ANY ($3) "+
			"ON CONFLICT ON CONSTRAINT lesson_members_states_pk DO "+
			"UPDATE SET updated_at = now(), %s WHERE lesson_members_states.state_type = $2 AND lesson_members_states.deleted_at IS NULL",
		strings.Join(fields, ", "), placeHolders, database.GenerateUpdatePlaceholders(fields, 4),
	)

	args := []interface{}{lessonID, stateType, userIds}
	args = append(args, values...)
	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *LessonMemberRepo) UpdateLessonMembersFields(ctx context.Context, db database.QueryExecer, e []*entities.LessonMember, updateFields entities.UpdateLessonMemberFields) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpdateLessonMembersFields")
	defer span.End()

	b := &pgx.Batch{}
	for _, item := range e {
		r.UpdateFieldsQueue(b, item, updateFields)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (r *LessonMemberRepo) GetLessonMembersInLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.LessonMembers, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMemberInLesson")
	defer span.End()

	fields, _ := (&entities.LessonMember{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_members
		WHERE lesson_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	members := entities.LessonMembers{}
	err := database.Select(ctx, db, query, &lessonID).ScanAll(&members)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return members, nil
}

func (r *LessonMemberRepo) GetLessonMembersInLessons(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) (entities.LessonMembers, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMembersInLessons")
	defer span.End()
	fields, _ := (&entities.LessonMember{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_members
		WHERE lesson_id = ANY($1) AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)
	members := entities.LessonMembers{}
	err := database.Select(ctx, db, query, &lessonIDs).ScanAll(&members)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return members, nil
}
