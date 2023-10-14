package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vl_payloads "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type LessonMemberRepo struct{}

type MemberStatesFilter struct {
	LessonID  pgtype.Text
	UserID    pgtype.Text
	StateType pgtype.Text
}

func (l *LessonMemberRepo) GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *MemberStatesFilter) (LessonMemberStateDTOs, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMemberStatesWithParams")
	defer span.End()
	fields, _ := (&LessonMemberStateDTO{}).FieldMap()
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
	states := LessonMemberStateDTOs{}
	err := database.Select(ctx, db, query, args...).ScanAll(&states)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return states, nil
}

func (l *LessonMemberRepo) UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *LessonMemberStateDTO) error {
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

func (l *LessonMemberRepo) UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID string, stateType domain.LearnerStateType, state *StateValueDTO) error {
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

func (l *LessonMemberRepo) UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID string, stateType domain.LearnerStateType, userIds []string, state *StateValueDTO) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpsertMultiLessonMemberStateByState")
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

func (l *LessonMemberRepo) GetLessonMembersInLesson(ctx context.Context, db database.QueryExecer, lessonID string) (LessonMemberDTOs, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMembersInLesson")
	defer span.End()

	fields, _ := (&LessonMemberDTO{}).FieldMap()
	query := fmt.Sprintf(`SELECT %s
		FROM lesson_members
		WHERE lesson_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
	)

	membersDTOs := LessonMemberDTOs{}
	err := database.Select(ctx, db, query, &lessonID).ScanAll(&membersDTOs)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return membersDTOs, nil
}

func (l *LessonMemberRepo) GetCourseAccessible(ctx context.Context, db database.QueryExecer, userID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetCourseAccessible")
	defer span.End()

	query := `SELECT l.course_id
		FROM lesson_members lm 
		JOIN lessons l ON lm.lesson_id = l.lesson_id
		AND lm.user_id = $1
		AND lm.deleted_at IS NULL
		AND l.deleted_at IS NULL`

	rows, err := db.Query(ctx, query, &userID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	studentCourseIDs := []string{}
	var courseID pgtype.Text
	for rows.Next() {
		if err := rows.Scan(&courseID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		studentCourseIDs = append(studentCourseIDs, courseID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return studentCourseIDs, nil
}

func (l *LessonMemberRepo) GetLessonMemberStatesByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (domain.LessonMemberStates, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMemberStatesByLessonID")
	defer span.End()

	filter := MemberStatesFilter{}
	errFilter := multierr.Combine(
		filter.LessonID.Set(lessonID),
		filter.UserID.Set(nil),
		filter.StateType.Set(nil),
	)
	if errFilter != nil {
		return nil, fmt.Errorf("error in setting up filter: %w", errFilter)
	}

	states, err := l.GetLessonMemberStatesWithParams(ctx, db, &filter)
	if err != nil {
		return nil, fmt.Errorf("error in GetLessonMemberStatesWithParams: %w", err)
	}

	return states.ToLessonMemberStatesDomainEntity(), nil
}

func (l *LessonMemberRepo) GetLearnerIDsByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLearnerIDsByLessonID")
	defer span.End()

	query := "SELECT lm.user_id FROM lesson_members lm where lm.lesson_id = $1 and lm.deleted_at IS NULL"
	rows, err := db.Query(ctx, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	var lessonMemberIDs []string
	for rows.Next() {
		var id pgtype.Text
		if err = rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		lessonMemberIDs = append(lessonMemberIDs, id.String)
	}
	return lessonMemberIDs, nil
}

func (l *LessonMemberRepo) GetLearnersByLessonIDWithPaging(ctx context.Context, db database.QueryExecer, params *vl_payloads.GetLearnersByLessonIDArgs) ([]domain.LessonMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLearnersByLessonIDWithPaging")
	defer span.End()

	dto := &LessonMemberDTO{}
	fields, values := dto.FieldMap()
	var rows pgx.Rows
	var err error
	query := fmt.Sprintf(`SELECT lm.%s
		FROM lesson_members AS lm
		WHERE lm.lesson_id = $1
		AND (($2::text IS NULL AND $3::text IS NULL) OR (concat(lm.lesson_id, lm.course_id), lm.user_id) > ($2::text COLLATE "C", $3::text)) 
		AND lm.deleted_at IS NULL
		ORDER BY concat(lm.lesson_id, lm.course_id) COLLATE "C" ASC, lm.user_id ASC
		LIMIT $4`,
		strings.Join(fields, ", lm."),
	)

	rows, err = db.Query(ctx, query, params.LessonID, params.LessonCourseID, params.UserID, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result []domain.LessonMember
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result = append(result, dto.ToLessonMemberDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return result, nil
}

func (l *LessonMemberRepo) GetLessonMemberUsersByLessonID(ctx context.Context, db database.QueryExecer, params *vl_payloads.GetLessonMemberUsersByLessonIDArgs) ([]*domain.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMemberUsersByLessonID")
	defer span.End()

	user := &User{}
	fields, values := user.FieldMap()
	var rows pgx.Rows
	var err error

	if params.UseLessonmgmtDB {
		query := fmt.Sprintf(`SELECT u.%s
			FROM users u
			WHERE u.user_id = ANY($1)
			AND u.deleted_at IS NULL`,
			strings.Join(fields, ", u."),
		)

		rows, err = db.Query(ctx, query, params.StudentIDs)
	} else {
		query := fmt.Sprintf(`SELECT u.%s
			FROM lesson_members AS lm
			INNER JOIN users AS u ON lm.user_id = u.user_id
			WHERE lm.lesson_id = $1
			AND u.user_group = 'USER_GROUP_STUDENT'
			AND lm.deleted_at IS NULL
			AND u.deleted_at IS NULL`,
			strings.Join(fields, ", u."),
		)

		rows, err = db.Query(ctx, query, params.LessonID)
	}
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result []*domain.User
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result = append(result, user.ToUserDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return result, nil
}

func (l *LessonMemberRepo) InsertMissingLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID string, stateType domain.LearnerStateType, state *StateValueDTO) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.InsertMissingLessonMemberStateByState")
	defer span.End()

	fields, values := state.FieldMap()
	placeHolders := database.GeneratePlaceholdersWithFirstIndex(3, len(fields)) // 2 first placeHolders is lesson_id and state_type
	query := fmt.Sprintf(
		"INSERT INTO lesson_members_states (lesson_id, user_id, state_type, updated_at, %s) "+
			"SELECT lesson_id , user_id , $2, now(), %s FROM lesson_members "+
			"WHERE lesson_id = $1 AND deleted_at IS NULL "+
			"ON CONFLICT ON CONSTRAINT lesson_members_states_pk DO NOTHING ",
		strings.Join(fields, ", "), placeHolders,
	)

	args := []interface{}{lessonID, stateType}
	args = append(args, values...)
	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonMemberRepo) InsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *LessonMemberStateDTO) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.InsertLessonMemberState")
	defer span.End()

	fields, args := state.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))
	query := fmt.Sprintf("INSERT INTO lesson_members_states (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT lesson_members_states_pk DO NOTHING ",
		strings.Join(fields, ", "), placeHolders)

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonMemberRepo) GetLessonLearnersByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) (map[string]domain.LessonLearners, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonLearnersByLessonIDs")
	defer span.End()

	dto := &LessonMemberDTO{}
	fields, values := dto.FieldMap()
	query := fmt.Sprintf(`SELECT %s
			FROM %s
			WHERE lesson_id = ANY($1)
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		dto.TableName(),
	)

	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	// fetch results of query
	lessonLearnersMap := make(map[string]domain.LessonLearners, len(lessonIDs))
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonLearner := &domain.LessonLearner{
			LearnerID: dto.UserID.String,
		}
		lessonLearnersMap[dto.LessonID.String] = append(lessonLearnersMap[dto.LessonID.String], lessonLearner)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonLearnersMap, nil
}
