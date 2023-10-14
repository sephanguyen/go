package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

// QuestionnaireRepo repo for questionnaire table
type QuestionnaireRepo struct{}

func (repo *QuestionnaireRepo) Upsert(ctx context.Context, db database.QueryExecer, questionnaire *entities.Questionnaire) error {
	now := time.Now()
	err := multierr.Combine(
		questionnaire.CreatedAt.Set(now),
		questionnaire.UpdatedAt.Set(now),
		questionnaire.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if questionnaire.QuestionnaireID.String == "" {
		_ = questionnaire.QuestionnaireID.Set(idutil.ULIDNow())
	}

	fields := database.GetFieldNames(questionnaire)
	values := database.GetScanFields(questionnaire, fields)
	placeHolders := database.GeneratePlaceholders(len(fields))
	tableName := questionnaire.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as qn (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__questionnaires 
		DO UPDATE SET 
			questionnaire_template_id = EXCLUDED.questionnaire_template_id,
			resubmit_allowed = EXCLUDED.resubmit_allowed,
			expiration_date = EXCLUDED.expiration_date,
			updated_at = EXCLUDED.updated_at
		WHERE qn.deleted_at IS NULL;
	`, tableName, strings.Join(fields, ", "), placeHolders)

	cmd, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("can not upsert questionnaire")
	}

	return nil
}

func (repo *QuestionnaireRepo) SoftDelete(ctx context.Context, db database.QueryExecer, questionnaireID []string) error {
	pgIDs := database.TextArray(questionnaireID)

	query := `
		UPDATE questionnaires AS qn
		SET deleted_at = now(), 
			updated_at = now() 
		WHERE questionnaire_id = ANY($1) 
		AND qn.deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, &pgIDs)
	if err != nil {
		return err
	}

	return nil
}

func (repo *QuestionnaireRepo) FindByID(ctx context.Context, db database.QueryExecer, id string) (*entities.Questionnaire, error) {
	var ent = &entities.Questionnaire{}
	fields := strings.Join(database.GetFieldNames(ent), ",")
	_, values := ent.FieldMap()
	err := db.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s
		FROM questionnaires qn
		WHERE qn.questionnaire_id=$1
		AND qn.deleted_at IS NULL
	`, fields), database.Text(id)).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow %w", err)
	}
	return ent, nil
}

func (repo *QuestionnaireRepo) FindQuestionsByQnID(ctx context.Context, db database.QueryExecer, id string) (entities.QuestionnaireQuestions, error) {
	fields := strings.Join(database.GetFieldNames(&entities.QuestionnaireQuestion{}), ",")
	ents := entities.QuestionnaireQuestions{}
	err := database.Select(ctx, db, fmt.Sprintf(`
		SELECT %s
		FROM questionnaire_questions qq
		WHERE qq.questionnaire_id=$1
		AND qq.deleted_at IS NULL
		ORDER BY qq.order_index ASC;
	`, fields), database.Text(id)).ScanAll(&ents)
	if err != nil {
		return nil, fmt.Errorf("database.Select %w", err)
	}
	return ents, nil
}

type FindUserAnswersFilter struct {
	QuestionnaireQuestionIDs pgtype.TextArray
	UserIDs                  pgtype.TextArray
	TargetIDs                pgtype.TextArray
	UserNotificationIDs      pgtype.TextArray
}

func NewFindUserAnswersFilter() FindUserAnswersFilter {
	f := FindUserAnswersFilter{}
	_ = f.QuestionnaireQuestionIDs.Set(nil)
	_ = f.UserIDs.Set(nil)
	_ = f.UserNotificationIDs.Set(nil)
	_ = f.TargetIDs.Set(nil)
	return f
}

func (repo *QuestionnaireRepo) FindUserAnswers(ctx context.Context, db database.QueryExecer, filter *FindUserAnswersFilter) (entities.QuestionnaireUserAnswers, error) {
	fields := strings.Join(database.GetFieldNames(&entities.QuestionnaireUserAnswer{}), ",")
	ents := entities.QuestionnaireUserAnswers{}
	err := database.Select(ctx, db, fmt.Sprintf(`
		SELECT %s
		FROM questionnaire_user_answers qn_ua
		WHERE ($1::TEXT[] IS NULL OR qn_ua.questionnaire_question_id=ANY($1))
			AND ($2::TEXT[] IS NULL OR qn_ua.user_notification_id=ANY($2))
			AND ($3::TEXT[] IS NULL OR qn_ua.user_id=ANY($3))
			AND ($4::TEXT[] IS NULL OR qn_ua.target_id=ANY($4))
			AND qn_ua.deleted_at IS NULL
	`, fields), filter.QuestionnaireQuestionIDs, filter.UserNotificationIDs, filter.UserIDs, filter.TargetIDs).ScanAll(&ents)
	if err != nil {
		return nil, fmt.Errorf("database.Select %w", err)
	}
	return ents, nil
}

type FindQuestionnaireRespondersFilter struct {
	UserName        pgtype.Text
	QuestionnaireID pgtype.Text
	Limit           pgtype.Int8
	Offset          pgtype.Int8
}

type QuestionnaireResponder struct {
	UserNotificationID pgtype.Text
	UserID             pgtype.Text
	IsParent           pgtype.Bool
	Name               pgtype.Text
	TargetID           pgtype.Text
	TargetName         pgtype.Text
	SubmittedAt        pgtype.Timestamptz
	IsIndividual       pgtype.Bool
}

func NewFindQuestionnaireRespondersFilter() FindQuestionnaireRespondersFilter {
	f := FindQuestionnaireRespondersFilter{}
	_ = f.QuestionnaireID.Set(nil)
	_ = f.UserName.Set("")
	_ = f.Limit.Set(nil)
	_ = f.Offset.Set(nil)
	return f
}

func (repo *QuestionnaireRepo) FindQuestionnaireResponders(ctx context.Context, db database.QueryExecer, filter *FindQuestionnaireRespondersFilter) (uint32, []*QuestionnaireResponder, error) {
	questionnaireResponders := []*QuestionnaireResponder{}
	query := `
		SELECT responders.user_notification_id, responders.user_id, responders.is_parent, responders.name, responders.target_id, responders.qn_submitted_at, responders.target_name, responders.is_individual
		FROM (
			SELECT u_ifn.user_notification_id, 
				u_ifn.user_id,
				CASE 
					WHEN student_id IS NOT NULL THEN student_id
					ELSE parent_id 
				END AS target_id,
				CASE 
					WHEN student_id IS NOT NULL AND parent_id IS NOT NULL THEN TRUE
					WHEN student_id IS NULL AND parent_id IS NOT NULL THEN TRUE
					ELSE FALSE
				END AS is_parent,
				CASE 
					WHEN user_id = student_id THEN student_name
					ELSE parent_name
				END AS name,
				u_ifn.qn_submitted_at,
				CASE
					WHEN student_id IS NOT NULL THEN student_name
					ELSE parent_name
				END AS target_name,
				is_individual
			FROM users_info_notifications u_ifn
				INNER JOIN info_notifications ifn ON u_ifn.notification_id = ifn.notification_id
			WHERE (
					(u_ifn.user_id = student_id AND u_ifn.student_name ILIKE CONCAT('%%', $1::TEXT, '%%')) OR
					(u_ifn.user_id = parent_id AND u_ifn.parent_name ILIKE CONCAT('%%', $1::TEXT, '%%'))  
				)
				AND ifn.questionnaire_id=$2
				AND u_ifn.qn_status='USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED'
				AND u_ifn.deleted_at IS NULL
				AND ifn.deleted_at IS NULL
			ORDER BY u_ifn.qn_submitted_at DESC
			LIMIT $3
			OFFSET $4
		) AS responders
	`

	rows, err := db.Query(ctx, query, filter.UserName, filter.QuestionnaireID, filter.Limit, filter.Offset)
	if err != nil {
		return 0, nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		qR := new(QuestionnaireResponder)
		if err := rows.Scan(&qR.UserNotificationID, &qR.UserID, &qR.IsParent, &qR.Name, &qR.TargetID, &qR.SubmittedAt, &qR.TargetName, &qR.IsIndividual); err != nil {
			return 0, nil, fmt.Errorf("rows.Scan: %w", err)
		}
		questionnaireResponders = append(questionnaireResponders, qR)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, fmt.Errorf("rows.Err(): %w", err)
	}

	var totalCount uint32

	queryCount := `
		SELECT count(*) AS total_count
		FROM (
			SELECT u_ifn.user_id
			FROM users_info_notifications u_ifn
				INNER JOIN info_notifications ifn ON u_ifn.notification_id = ifn.notification_id
			WHERE (
					(u_ifn.user_id = student_id AND u_ifn.student_name ILIKE CONCAT('%%', $1::TEXT, '%%')) OR
					(u_ifn.user_id = parent_id AND u_ifn.parent_name ILIKE CONCAT('%%', $1::TEXT, '%%'))  
				)
				AND ifn.questionnaire_id=$2
				AND u_ifn.qn_status='USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED'
				AND u_ifn.deleted_at IS null
				AND ifn.deleted_at IS NULL
		) AS responders
	`
	row := db.QueryRow(ctx, queryCount, filter.UserName, filter.QuestionnaireID)
	err = row.Scan(&totalCount)
	if err != nil {
		return 0, nil, fmt.Errorf("row.Scan: %w", err)
	}

	return totalCount, questionnaireResponders, nil
}

type QuestionnaireCSVResponder struct {
	UserNotificationID pgtype.Text
	UserID             pgtype.Text
	IsParent           pgtype.Bool
	Name               pgtype.Text
	TargetID           pgtype.Text
	StudentID          pgtype.Text
	TargetName         pgtype.Text
	SubmittedAt        pgtype.Timestamptz
	IsIndividual       pgtype.Bool
	StudentExternalID  pgtype.Text
	LocationNames      pgtype.TextArray
	SubmissionStatus   pgtype.Text
}

func (repo *QuestionnaireRepo) FindQuestionnaireCSVResponders(ctx context.Context, db database.QueryExecer, questionnaireID string) ([]*QuestionnaireCSVResponder, error) {
	questionnaireResponders := []*QuestionnaireCSVResponder{}
	query := `
		SELECT responders.user_notification_id, 
			responders.user_id, 
			responders.is_parent, 
			responders.name, 
			responders.target_id, 
			responders.student_id,
			responders.qn_submitted_at, 
			responders.target_name, 
			responders.is_individual,
			responders.student_external_id,
			responders.location_names,
			responders.submitted_status
		FROM (
			SELECT u_ifn.user_notification_id, 
				u_ifn.user_id,
				CASE 
					WHEN u_ifn.student_id IS NOT NULL THEN u_ifn.student_id
					ELSE parent_id 
				END AS target_id,
				CASE 
					WHEN u_ifn.student_id IS NOT NULL AND parent_id IS NOT NULL THEN TRUE
					WHEN u_ifn.student_id IS NULL AND parent_id IS NOT NULL THEN TRUE
					ELSE FALSE
				END AS is_parent,
				CASE 
					WHEN u_ifn.user_id = u_ifn.student_id THEN student_name
					ELSE parent_name
				END AS name,
				u_ifn.qn_submitted_at,
				CASE
					WHEN u_ifn.student_id IS NOT NULL THEN student_name
					ELSE parent_name
				END AS target_name,
				is_individual,
				s.student_external_id,
				s.student_id,
				array_agg(l."name") location_names,
				u_ifn.qn_status submitted_status
			FROM users_info_notifications u_ifn
				INNER JOIN info_notifications ifn ON u_ifn.notification_id = ifn.notification_id
				LEFT JOIN students s ON s.student_id = u_ifn.student_id
				LEFT JOIN user_access_paths uap ON u_ifn.user_id = uap.user_id 
				LEFT JOIN locations l ON uap.location_id = l.location_id 
			WHERE ifn.questionnaire_id=$1
				AND u_ifn.deleted_at IS NULL
				AND ifn.deleted_at IS NULL
				AND s.deleted_at IS NULL 
				AND uap.deleted_at IS NULL 
				AND l.deleted_at IS NULL 
			GROUP BY (u_ifn.user_notification_id, u_ifn.user_id, s.student_external_id, s.student_id)
			ORDER BY u_ifn.qn_submitted_at DESC NULLS LAST
		) AS responders
	`

	rows, err := db.Query(ctx, query, database.Text(questionnaireID))
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		qR := new(QuestionnaireCSVResponder)
		if err := rows.Scan(&qR.UserNotificationID, &qR.UserID, &qR.IsParent, &qR.Name, &qR.TargetID, &qR.StudentID, &qR.SubmittedAt, &qR.TargetName, &qR.IsIndividual, &qR.StudentExternalID, &qR.LocationNames, &qR.SubmissionStatus); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		questionnaireResponders = append(questionnaireResponders, qR)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	return questionnaireResponders, nil
}

// QuestionnaireQuestionRepo repo for questionnaire_questions table
type QuestionnaireQuestionRepo struct{}

// ForceUpsert -> set deleted_at = nil
func (repo *QuestionnaireQuestionRepo) queueForceUpsert(b *pgx.Batch, item *entities.QuestionnaireQuestion) error {
	now := time.Now()
	err := multierr.Combine(
		item.CreatedAt.Set(now),
		item.UpdatedAt.Set(now),
		item.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if item.QuestionnaireQuestionID.String == "" {
		_ = item.QuestionnaireQuestionID.Set(idutil.ULIDNow())
	}

	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as qnq (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__questionnaire_questions
		DO UPDATE SET
			questionnaire_id = EXCLUDED.questionnaire_id,
			order_index = EXCLUDED.order_index,
			type = EXCLUDED.type,
			title = EXCLUDED.title,
			choices = EXCLUDED.choices,
			is_required = EXCLUDED.is_required,
			updated_at = EXCLUDED.updated_at,
			deleted_at = NULL;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *QuestionnaireQuestionRepo) BulkForceUpsert(ctx context.Context, db database.QueryExecer, items entities.QuestionnaireQuestions) error {
	b := &pgx.Batch{}
	for _, item := range items {
		err := repo.queueForceUpsert(b, item)
		if err != nil {
			return fmt.Errorf("repo.queueForceUpsert: %w", err)
		}
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (repo *QuestionnaireQuestionRepo) SoftDelete(ctx context.Context, db database.QueryExecer, questionnaireID []string) error {
	pgIDs := database.TextArray(questionnaireID)

	query := `
		UPDATE questionnaire_questions as qn_q
		SET deleted_at = now(), 
			updated_at = now() 
		WHERE questionnaire_id = ANY($1) 
		AND qn_q.deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, &pgIDs)
	if err != nil {
		return err
	}

	return nil
}

// QuestionnaireQuestionRepo repo for questionnaire_user_answers table
type QuestionnaireUserAnswerRepo struct{}

func (repo *QuestionnaireUserAnswerRepo) SoftDelete(ctx context.Context, db database.QueryExecer, answerIDs []string) error {
	pgIDs := database.TextArray(answerIDs)

	query := `
		UPDATE questionnaire_user_answers as qn_ua
		SET deleted_at = now()
		WHERE answer_id = ANY($1) 
		AND qn_ua.deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, &pgIDs)
	if err != nil {
		return err
	}

	return nil
}

func (repo *QuestionnaireUserAnswerRepo) queueUpsert(b *pgx.Batch, item *entities.QuestionnaireUserAnswer) error {
	err := multierr.Combine(
		item.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if item.AnswerID.String == "" {
		_ = item.AnswerID.Set(idutil.ULIDNow())
	}

	fieldNames := database.GetFieldNames(item)
	values := database.GetScanFields(item, fieldNames)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	tableName := item.TableName()

	query := fmt.Sprintf(`
		INSERT INTO %s as qn_ua (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT pk__questionnaire_user_answers 
		DO NOTHING;
	`, tableName, strings.Join(fieldNames, ", "), placeHolders)

	b.Queue(query, values...)

	return nil
}

func (repo *QuestionnaireUserAnswerRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items entities.QuestionnaireUserAnswers) error {
	b := &pgx.Batch{}
	for _, item := range items {
		err := repo.queueUpsert(b, item)
		if err != nil {
			return fmt.Errorf("repo.queueForceUpsert: %w", err)
		}
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (repo *QuestionnaireUserAnswerRepo) SoftDeleteByQuestionnaireID(ctx context.Context, db database.QueryExecer, questionnaireID []string) error {
	ctx, span := interceptors.StartSpan(ctx, "SoftDeleteByQuestionnaireID")
	defer span.End()

	query := `
		UPDATE questionnaire_user_answers AS qn_ua
		SET deleted_at = now()
		FROM questionnaires q
		JOIN questionnaire_questions qq ON qq.questionnaire_id = q.questionnaire_id 
		WHERE qq.questionnaire_question_id = qn_ua.questionnaire_question_id 
		AND q.questionnaire_id = ANY($1)
		AND qn_ua.deleted_at IS NULL;
	`

	_, err := db.Exec(ctx, query, database.TextArray(questionnaireID))
	if err != nil {
		return err
	}

	return nil
}
