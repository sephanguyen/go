package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QuizRepo works with quizs
type QuizRepo struct{}

const quizRepoSearchStmt = `SELECT %s FROM quizzes
WHERE deleted_at IS NULL
AND external_id = ANY($1)
AND ($2::text IS NULL OR status = $2)
LIMIT %d`

const quizRepoDeleteByExternalIDStmt = `UPDATE quizzes
SET deleted_at = NOW(), status = 'QUIZ_STATUS_DELETED'
WHERE external_id = $1 AND school_id = $2 AND deleted_at IS NULL`

// Create creates Quiz
func (r *QuizRepo) Create(ctx context.Context, db database.QueryExecer, quiz *entities.Quiz) error {
	now := timeutil.Now()
	err := multierr.Combine(
		quiz.ID.Set(idutil.ULID(now)),
		quiz.CreatedAt.Set(now),
		quiz.UpdatedAt.Set(now),
		quiz.DeletedAt.Set(nil),
	)
	if err != nil {
		return err
	}

	cmd, err := database.Insert(ctx, quiz, db.Exec)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("can not create quiz")
	}

	return nil
}

func (r *QuizRepo) Upsert(ctx context.Context, db database.QueryExecer, data []*entities.Quiz) ([]*entities.Quiz, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuizRepo.Upsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, t *entities.Quiz) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`
			INSERT INTO quizzes(%s) VALUES(%s)
				ON CONFLICT ON CONSTRAINT quizs_pk
			DO UPDATE SET
				country = $2,
				school_id = $3,
				lo_ids = $4,
				external_id = $5,
				kind = $6,
				question = $7,
				explanation = $8,
				options = $9,
				tagged_los = $10,
				difficulty_level = $11,
				point = $12,
				question_tag_ids = $13,
				created_by = $14,
				approved_by = $15,
				status = $16,
				updated_at = $17,
				created_at = $18,
				deleted_at = $19,
				label_type = $21
			RETURNING %s
		`,
			strings.Join(fieldNames, ","),
			placeHolders,
			strings.Join(fieldNames, ","),
		)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, each := range data {
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	resp := make([]*entities.Quiz, 0)
	for i := 0; i < b.Len(); i++ {
		quiz := new(entities.Quiz)
		_, values := quiz.FieldMap()
		if err := result.QueryRow().Scan(values...); err != nil {
			return nil, fmt.Errorf("batchResults.QueryRow: %w", err)
		}
		resp = append(resp, quiz)
	}

	return resp, nil
}

// GetByExternalID retrieves a single Quiz by external id
func (r *QuizRepo) GetByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) (*entities.Quiz, error) {
	e := &entities.Quiz{}

	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
			FROM %s
			WHERE deleted_at is NULL
			AND external_id = $1::TEXT
			AND school_id = $2::INT4`,
		strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, stmt, &id, &schoolID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

// QuizFilter can be use with Search
type QuizFilter struct {
	ExternalIDs pgtype.TextArray
	Status      pgtype.Text
	Limit       uint
}

// Search returns by user ID
func (r *QuizRepo) Search(ctx context.Context, db database.QueryExecer, filter QuizFilter) (entities.Quizzes, error) {
	e := &entities.Quiz{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(quizRepoSearchStmt, strings.Join(fields, ","), filter.Limit)

	results := make(entities.Quizzes, 0, filter.Limit)
	err := database.Select(ctx, db, stmt,
		&filter.ExternalIDs,
		&filter.Status).ScanAll(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetByExternalIDs Retrieve retrieves a multiple Quiz by external id
func (r *QuizRepo) GetByExternalIDsAndLOID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, loID pgtype.Text) (entities.Quizzes, error) {
	e := &entities.Quiz{}
	loe := &entities.LearningObjective{}

	fields, _ := e.FieldMap()
	for i := range fields {
		fields[i] = "q." + fields[i]
	}
	stmt := fmt.Sprintf(`SELECT %s
			FROM %s q INNER JOIN %s lo ON q.school_id = lo.school_id
			INNER JOIN unnest($1::TEXT[]) WITH ORDINALITY AS search_quiz_external_ids(id, id_order)
			ON q.external_id = search_quiz_external_ids.id
			WHERE q.deleted_at is NULL
			AND lo.lo_id = $2
			ORDER BY search_quiz_external_ids.id_order;`,
		strings.Join(fields, ","), e.TableName(), loe.TableName())

	quizzes := entities.Quizzes{}
	if err := database.Select(ctx, db, stmt, &ids, &loID).ScanAll(&quizzes); err != nil {
		return nil, err
	}

	return quizzes, nil
}

// GetByExternalIDs Retrieve retrieves a multiple Quiz by external id
func (r *QuizRepo) GetByExternalIDsAndLmID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, lmID pgtype.Text) (entities.Quizzes, error) {
	e := &entities.Quiz{}

	fields, _ := e.FieldMap()
	for i := range fields {
		fields[i] = "q." + fields[i]
	}
	stmt := fmt.Sprintf(
		`SELECT %s FROM %s q
			INNER JOIN UNNEST($1::TEXT[]) WITH ORDINALITY AS external_ids(id, id_order) ON q.external_id = external_ids.id
		WHERE
			q.deleted_at is NULL
			AND lo_ids[1] = $2
		ORDER BY
			external_ids.id_order;`,
		strings.Join(fields, ","), e.TableName(),
	)

	quizzes := entities.Quizzes{}
	if err := database.Select(ctx, db, stmt, &ids, &lmID).ScanAll(&quizzes); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return quizzes, nil
}

func (r *QuizRepo) GetByQuestionGroupID(ctx context.Context, db database.QueryExecer, questionGroupID pgtype.Text) (entities.Quizzes, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuizRepo.GetByQuestionGroupID")
	defer span.End()

	e := &entities.Quiz{}

	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
			FROM %s
			WHERE deleted_at is NULL
			AND question_group_id = $1::TEXT`,
		strings.Join(fields, ","), e.TableName())
	quizzes := entities.Quizzes{}

	if err := database.Select(ctx, db, stmt, &questionGroupID).ScanAll(&quizzes); err != nil {
		return nil, err
	}
	return quizzes, nil
}

// QuizSetFilter for search
type QuizSetFilter struct {
	ObjectiveIDs pgtype.TextArray
	Status       pgtype.Text
	Limit        uint
}

// QuizSetRepo works with quiz_sets

const quizSetRepoSearchStmt = `SELECT %s FROM quiz_sets
WHERE deleted_at IS NULL
AND ($1::text[] IS NULL OR lo_id = ANY ($1))
AND ($2::text IS NULL OR status = $2)
ORDER BY quiz_set_id DESC
LIMIT %d`

const quizSetRepoDeleteStmt = `UPDATE quiz_sets
SET deleted_at = NOW(), status = 'QUIZSET_STATUS_DELETED'
WHERE quiz_set_id = $1
AND deleted_at IS NULL`

// Search searches quiz_sets match the filter
func (r *QuizSetRepo) Search(ctx context.Context, db database.QueryExecer, filter QuizSetFilter) (entities.QuizSets, error) {
	if len(filter.ObjectiveIDs.Elements) == 0 {
		return nil, nil
	}

	e := &entities.QuizSet{}
	stmt := fmt.Sprintf(quizSetRepoSearchStmt, strings.Join(database.GetFieldNames(e), ","), filter.Limit)

	results := make(entities.QuizSets, 0, filter.Limit)
	err := database.Select(ctx, db, stmt,
		&filter.ObjectiveIDs,
		&filter.Status).ScanAll(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// Delete with soft delete
func (r *QuizSetRepo) Delete(ctx context.Context, db database.QueryExecer, id pgtype.Text) error {
	cmd, err := db.Exec(ctx, quizSetRepoDeleteStmt, &id)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any quizset to delete: %w", pgx.ErrNoRows)
	}

	return nil
}

// Create inserts new quiz_set
func (r *QuizSetRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.QuizSet) error {
	now := timeutil.Now()
	err := multierr.Combine(
		e.ID.Set(idutil.ULID(now)),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.DeletedAt.Set(nil),
	)
	if err != nil {
		return err
	}

	cmd, err := database.Insert(ctx, e, db.Exec)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("can not create quizset")
	}

	return nil
}

// GetQuizSetByLoID get quizset by learning object id
func (r *QuizSetRepo) GetQuizSetByLoID(ctx context.Context, db database.QueryExecer, loID pgtype.Text) (*entities.QuizSet, error) {
	quizSet := &entities.QuizSet{}

	field, _ := quizSet.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND lo_id = $1;`, strings.Join(field, ","), quizSet.TableName())

	if err := database.Select(ctx, db, stmt, loID).ScanOne(quizSet); err != nil {
		return nil, err
	}

	return quizSet, nil
}

// GetQuizSetsOfLOContainQuiz get all quiz set belong to LO that contains the quiz id
func (r *QuizSetRepo) GetQuizSetsOfLOContainQuiz(ctx context.Context, db database.QueryExecer, loID pgtype.Text, quizID pgtype.Text) (entities.QuizSets, error) {
	quizSets := entities.QuizSets{}
	quizSet := entities.QuizSet{}
	field, _ := quizSet.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND lo_id = $1
	AND quiz_external_ids @> array[$2];`, strings.Join(field, ","), quizSet.TableName())

	if err := database.Select(ctx, db, stmt, loID, quizID).ScanAll(&quizSets); err != nil {
		return nil, err
	}

	return quizSets, nil
}

// GetQuizSetsContainQuiz get all quizsets contain the quiz id
func (r *QuizSetRepo) GetQuizSetsContainQuiz(ctx context.Context, db database.QueryExecer, quizID pgtype.Text) (entities.QuizSets, error) {
	quizSets := entities.QuizSets{}
	quizSet := entities.QuizSet{}
	field, _ := quizSet.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND quiz_external_ids @> array[$1];`, strings.Join(field, ","), quizSet.TableName())

	if err := database.Select(ctx, db, stmt, quizID).ScanAll(&quizSets); err != nil {
		return nil, err
	}

	return quizSets, nil
}

// GetQuizExternalIDs returns quiz external ids of a quiz set
func (r *QuizSetRepo) GetQuizExternalIDs(ctx context.Context, db database.QueryExecer, loID pgtype.Text, limit pgtype.Int8, offset pgtype.Int8) ([]string, error) {
	quizSet := &entities.QuizSet{}

	stmt := fmt.Sprintf(`
		SELECT UNNEST(qs.quiz_external_ids) AS quiz_external_id FROM %v qs
		WHERE lo_id = $1
		AND deleted_at IS NULL
		LIMIT $2
		OFFSET $3
	`, quizSet.TableName())

	quizExternalIDs := make([]string, 0)
	rows, err := db.Query(ctx, stmt, loID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var id pgtype.Text
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		quizExternalIDs = append(quizExternalIDs, id.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return quizExternalIDs, nil
}

// GetOptions get options of a quiz by its ID
func (r *QuizRepo) GetOptions(ctx context.Context, db database.QueryExecer, quizID pgtype.Text, loID pgtype.Text) ([]*entities.QuizOption, error) {
	e := &entities.Quiz{}
	loe := &entities.LearningObjective{}
	optionsJSONB := database.JSONB("")
	stmt := fmt.Sprintf(`SELECT options
			FROM %s LEFT JOIN %s ON %s.school_id = %s.school_id
			WHERE %s.deleted_at is NULL
			AND external_id = $1::TEXT
			AND %s.lo_id=$2::TEXT;`,
		e.TableName(), loe.TableName(), e.TableName(), loe.TableName(), e.TableName(), loe.TableName(),
	)

	if err := db.QueryRow(ctx, stmt, quizID, loID).Scan(&optionsJSONB); err != nil {
		return nil, err
	}

	options := make([]*entities.QuizOption, 0)
	err := optionsJSONB.AssignTo(&options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

const quizRepoRetrieveStmt = `SELECT %s FROM quizzes
WHERE quiz_id = $1::TEXT`

// Retrieve retrieves a single Quiz
func (r *QuizRepo) Retrieve(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Quiz, error) {
	e := &entities.Quiz{}
	fields, _ := e.FieldMap()

	stmt := fmt.Sprintf(quizRepoRetrieveStmt, strings.Join(fields, ","))
	if err := database.Select(ctx, db, stmt, &id).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

// DeleteByExternalID with soft delete
func (r *QuizRepo) DeleteByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) error {
	cmd, err := db.Exec(ctx, quizRepoDeleteByExternalIDStmt, &id, &schoolID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any quiz to delete: %w", pgx.ErrNoRows)
	}

	return nil
}

func (r *QuizRepo) DeleteByQuestionGroupID(ctx context.Context, db database.QueryExecer, questionGroupID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "QuizRepo.DeleteByQuestionGroupID")
	defer span.End()

	cmd, err := db.Exec(ctx,
		`UPDATE quizzes
		SET deleted_at = NOW(), status = 'QUIZ_STATUS_DELETED'
		WHERE question_group_id = $1 AND deleted_at IS NULL`,
		&questionGroupID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("quiz not found: %w", pgx.ErrNoRows)
	}

	return nil
}

// QuizSetRepo works with quiz_sets
type QuizSetRepo struct{}

// GetTotalQuiz return total quiz array of lo_ids
func (r *QuizSetRepo) GetTotalQuiz(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error) {
	e := &entities.QuizSet{}

	res := make(map[string]int32)

	stmt := fmt.Sprintf(`SELECT lo_id, array_length(quiz_external_ids, 1)
	FROM %s
	WHERE deleted_at IS NULL AND lo_id = ANY($1::_TEXT)`, e.TableName())

	rows, err := db.Query(ctx, stmt, loIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id pgtype.Text
		var totalQuiz pgtype.Int4
		err := rows.Scan(&id, &totalQuiz)
		if err != nil {
			return nil, err
		}
		res[id.String] = totalQuiz.Int
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return res, nil
}

// Get total points from quiz set
func (r *QuizSetRepo) GetTotalPointsByQuizSetID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Int4, error) {
	e := entities.QuizSet{}

	totalPoints := database.Int4(0)
	// make sure to typecast the calculated value to int as this will return bigint which is not the type for the total_points column on exam_lo_submission
	query := fmt.Sprintf(`
	SELECT COALESCE((
				SELECT SUM(point)
				FROM public.quizzes q
				WHERE
					q.deleted_at IS NULL
					AND q.external_id = ANY(qs.quiz_external_ids)
			),
			0
		)::INT AS total_points
	FROM
		%s qs
	WHERE
		quiz_set_id = $1`, e.TableName())

	err := database.Select(ctx, db, query, id).ScanFields(&totalPoints)
	if err != nil {
		return totalPoints, err
	}

	return totalPoints, nil
}

func (r *QuizSetRepo) CountQuizOnLO(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionSetRepo.CountQuizOnLO")
	defer span.End()

	e := &entities.QuizSet{}

	query := fmt.Sprintf("SELECT lo_id, array_length(quiz_external_ids, 1) FROM %s WHERE lo_id = ANY($1) AND status = $2", e.TableName())

	rows, err := db.Query(ctx, query, &loIDs, "QUIZSET_STATUS_APPROVED")
	if err != nil {
		return nil, fmt.Errorf("repo.DB.QueryEx: %w", err)
	}
	defer rows.Close()

	ss := make(map[string]int32)

	for rows.Next() {
		var loID pgtype.Text
		var count pgtype.Int4
		if err := rows.Scan(&loID, &count); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		// only choose an quiz_external_ids not null
		if count.Status == pgtype.Present {
			ss[loID.String] = count.Int
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return ss, nil
}

func (r *QuizSetRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, data []*entities.QuizSet) ([]*entities.QuizSet, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuizSetRepo.BulkUpsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, t *entities.QuizSet) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`
			INSERT INTO %s(%s) VALUES(%s)
				ON CONFLICT ON CONSTRAINT quiz_sets_pk
			DO UPDATE SET
				status = $4,
				updated_at = NOW(),
				deleted_at = $6
			RETURNING %s
		`,
			t.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
			strings.Join(fieldNames, ","),
		)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	b := &pgx.Batch{}

	for _, each := range data {
		queueFn(b, each)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	resp := make([]*entities.QuizSet, 0)
	for i := 0; i < b.Len(); i++ {
		quiz := new(entities.QuizSet)
		_, values := quiz.FieldMap()
		if err := result.QueryRow().Scan(values...); err != nil {
			return nil, fmt.Errorf("batchResults.QueryRow: %w", err)
		}
		resp = append(resp, quiz)
	}

	return resp, nil
}

const quizSetRepoRetrieveByLoIDsStmt = `SELECT %s FROM %s
WHERE deleted_at IS NULL
AND ($1::text[] IS NULL OR lo_id = ANY ($1))
AND ($2::text IS NULL OR status = $2)
ORDER BY quiz_set_id DESC
`

// Search searches quiz_sets match the filter
func (r *QuizSetRepo) RetrieveByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (entities.QuizSets, error) {
	e := &entities.QuizSet{}
	stmt := fmt.Sprintf(quizSetRepoRetrieveByLoIDsStmt, strings.Join(database.GetFieldNames(e), ","), e.TableName())

	var results entities.QuizSets
	err := database.Select(ctx, db, stmt,
		&loIDs,
		pb.QuizSetStatus_QUIZSET_STATUS_APPROVED.String()).ScanAll(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ShuffledQuizSetRepo quizzes in the quiz set will be shuffled
type ShuffledQuizSetRepo struct{}

// Create inserts new quiz_set
func (r *ShuffledQuizSetRepo) Create(ctx context.Context, db database.QueryExecer, shuffledQuizSet *entities.ShuffledQuizSet) (pgtype.Text, error) {
	cmd, err := database.Insert(ctx, shuffledQuizSet, db.Exec)
	if err != nil {
		return pgtype.Text{Status: pgtype.Null}, err
	}

	if cmd.RowsAffected() != 1 {
		return pgtype.Text{Status: pgtype.Null}, fmt.Errorf("can not create shuffled quizset")
	}

	return shuffledQuizSet.ID, nil
}

// Get return shuffledQuizSet by id
func (r *ShuffledQuizSetRepo) Get(ctx context.Context, db database.QueryExecer, id pgtype.Text, from pgtype.Int8, to pgtype.Int8) (*entities.ShuffledQuizSet, error) {
	shuffledQuizSet := &entities.ShuffledQuizSet{}

	stmt := fmt.Sprintf(`SELECT shuffled_quiz_set_id, original_quiz_set_id, quiz_external_ids[$2:$3], status, random_seed, updated_at, created_at, deleted_at, student_id, total_correctness, submission_history, original_shuffle_quiz_set_id
	FROM %s
	WHERE deleted_at IS NULL
	AND shuffled_quiz_set_id = $1;`, shuffledQuizSet.TableName())

	if err := database.Select(ctx, db, stmt, id, from.Get(), to.Get()).ScanOne(shuffledQuizSet); err != nil {
		return nil, err
	}
	return shuffledQuizSet, nil
}

// GetSeed return seed of the shuffled quiz set use
func (r *ShuffledQuizSetRepo) GetSeed(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Text, error) {
	shuffledQuizSet := &entities.ShuffledQuizSet{}

	seed := pgtype.Text{}
	stmt := fmt.Sprintf(`SELECT random_seed
	FROM %s
	WHERE deleted_at IS NULL
	AND shuffled_quiz_set_id = $1;`, shuffledQuizSet.TableName())

	err := db.QueryRow(ctx, stmt, id).Scan(&seed)
	if err != nil {
		seed.Set(nil)
		return seed, err
	}

	return seed, nil
}

// UpdateSubmissionHistory this will update history answer of doing quiz
func (r *ShuffledQuizSetRepo) UpdateSubmissionHistory(ctx context.Context, db database.QueryExecer, id pgtype.Text, answer pgtype.JSONB) error {
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	stmt := fmt.Sprintf(`UPDATE %s SET submission_history = submission_history || $1, updated_at = now() WHERE shuffled_quiz_set_id = $2;
	`, shuffledQuizSet.TableName())

	cmd, err := db.Exec(ctx, stmt, answer, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any quizset to update submission_history: %w", pgx.ErrNoRows)
	}

	return nil
}

// UpdateTotalCorrectness will update total correct quiz of the test
func (r *ShuffledQuizSetRepo) UpdateTotalCorrectness(ctx context.Context, db database.QueryExecer, id pgtype.Text) error {
	shuffledQuizSet := &entities.ShuffledQuizSet{}

	stmt := fmt.Sprintf(`UPDATE %s SET total_correctness =
	(
		SELECT COUNT(DISTINCT (value->>'quiz_id')) AS quiz_id
		FROM %s CROSS JOIN jsonb_array_elements(submission_history)
		WHERE shuffled_quiz_set_id = $1 AND (value->>'is_accepted')::boolean = true
	),
	updated_at = NOW()
	WHERE shuffled_quiz_set_id = $1;`, shuffledQuizSet.TableName(), shuffledQuizSet.TableName())

	cmd, err := db.Exec(ctx, stmt, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetByStudyPlanItems will return the list of shuffled quiz set of studyPlanItem
func (r *ShuffledQuizSetRepo) GetByStudyPlanItems(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) (entities.ShuffledQuizSets, error) {
	ents := entities.ShuffledQuizSets{}
	e := entities.ShuffledQuizSet{}

	fieldNames := database.GetFieldNames(&e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_item_id = ANY($1)", strings.Join(fieldNames, ","), e.TableName())

	err := database.Select(ctx, db, query, studyPlanItemIDs).ScanAll(&ents)
	if err != nil {
		return ents, err
	}

	return ents, nil
}

// GetBySessionID will return the list of shuffled quiz set of session_id
func (r *ShuffledQuizSetRepo) GetBySessionID(ctx context.Context, db database.QueryExecer, sessionID pgtype.Text) (*entities.ShuffledQuizSet, error) {
	e := &entities.ShuffledQuizSet{}

	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE session_id = $1 and deleted_at is NULL ORDER BY created_at desc LIMIT 1", strings.Join(fieldNames, ","), e.TableName())

	err := database.Select(ctx, db, query, sessionID).ScanOne(e)
	if err != nil {
		return e, err
	}

	return e, nil
}

func (r *ShuffledQuizSetRepo) GetByStudyPlanItemIdentities(ctx context.Context, db database.QueryExecer, identities []*StudyPlanItemIdentity) (entities.ShuffledQuizSets, error) {
	ctx, span := interceptors.StartSpan(ctx, "ShuffledQuizSetRepo.GetByStudyPlanItemIdentities")
	defer span.End()

	ents := entities.ShuffledQuizSets{}
	e := entities.ShuffledQuizSet{}

	studentIDs, studyPlanIDs, lOIDs := make([]string, len(identities)), make([]string, len(identities)), make([]string, len(identities))
	for i, identity := range identities {
		studentIDs[i] = fmt.Sprintf(`'%s'`, identity.StudentID.String)
		studyPlanIDs[i] = fmt.Sprintf(`'%s'`, identity.StudyPlanID.String)
		lOIDs[i] = fmt.Sprintf(`'%s'`, identity.LearningMaterialID.String)
	}

	query := fmt.Sprintf(`
		SELECT sqs.%s
		FROM UNNEST(
				ARRAY[%s],
				ARRAY[%s],
				ARRAY[%s]
			) WITH ORDINALITY AS un(
				student_id, study_plan_id, learning_material_id
			)
		JOIN %s sqs USING(student_id, study_plan_id, learning_material_id)
		WHERE sqs.deleted_at IS NULL
	`,
		strings.Join(database.GetFieldNames(&e), ", sqs."),
		strings.Join(studentIDs, ", "),
		strings.Join(studyPlanIDs, ", "),
		strings.Join(lOIDs, ", "),
		e.TableName(),
	)

	// Execute
	if err := database.Select(ctx, db, query).ScanAll(&ents); err != nil {
		return nil, err
	}

	return ents, nil
}

const getSubmissionHistoryQuery = `
SELECT value, quiz_id
FROM %s sqs 
	INNER JOIN UNNEST(sqs.quiz_external_ids) WITH ORDINALITY AS quiz_id ON sqs.shuffled_quiz_set_id = $1 
	LEFT JOIN 
	(
		SELECT DISTINCT ON (value->>'quiz_id') value
		FROM %s sqs2
			INNER JOIN JSONB_ARRAY_ELEMENTS(sqs2.submission_history) ON shuffled_quiz_set_id = $1
		ORDER BY value->>'quiz_id', value->>'submitted_at' DESC
	) AS sub
	ON quiz_id = sub.value->>'quiz_id'
ORDER BY ORDINALITY
LIMIT $2 OFFSET $3;`

// GetSubmissionHistory will return student doing test history
func (r *ShuffledQuizSetRepo) GetSubmissionHistory(ctx context.Context, db database.QueryExecer, id pgtype.Text, limit, offset pgtype.Int4) (map[pgtype.Text]pgtype.JSONB, []pgtype.Text, error) {
	e := entities.ShuffledQuizSet{}

	// this query means that we will fetch the row distinct by quiz_id with latest submitted
	query := fmt.Sprintf(getSubmissionHistoryQuery, e.TableName(), e.TableName())

	rows, err := db.Query(ctx, query, id, limit.Get(), offset.Get())
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	mp := make(map[pgtype.Text]pgtype.JSONB)
	var orderedQuizList []pgtype.Text
	for rows.Next() {
		var quizID pgtype.Text
		var sub pgtype.JSONB
		err := rows.Scan(&sub, &quizID)
		if err != nil {
			return nil, nil, err
		}
		orderedQuizList = append(orderedQuizList, quizID)
		mp[quizID] = sub
	}
	return mp, orderedQuizList, nil
}

// GetStudentID returns the student id of the shuffled quiz set
func (r *ShuffledQuizSetRepo) GetStudentID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Text, error) {
	e := entities.ShuffledQuizSet{}

	studentID := pgtype.Text{}
	query := fmt.Sprintf(`SELECT student_id FROM %s WHERE shuffled_quiz_set_id = $1`, e.TableName())

	err := database.Select(ctx, db, query, id).ScanFields(&studentID)
	if err != nil {
		return studentID, err
	}
	return studentID, nil
}

// GetLoID returns the lo ID of the quiz set
func (r *ShuffledQuizSetRepo) GetLoID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Text, error) {
	shuffledQuizSet := entities.ShuffledQuizSet{}
	quizSet := entities.QuizSet{}

	loID := pgtype.Text{}
	query := fmt.Sprintf(`SELECT lo_id
	FROM %s INNER JOIN %s ON original_quiz_set_id = quiz_set_id
	WHERE shuffled_quiz_set_id = $1;`, shuffledQuizSet.TableName(), quizSet.TableName())

	err := database.Select(ctx, db, query, id).ScanFields(&loID)
	if err != nil {
		return loID, err
	}
	return loID, nil
}

// GetScore returns the total_correctness/total_quiz
func (r *ShuffledQuizSetRepo) GetScore(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Int4, pgtype.Int4, error) {
	e := entities.ShuffledQuizSet{}

	totalCorrectness := database.Int4(0)
	totalQuiz := database.Int4(0)
	query := fmt.Sprintf(`SELECT total_correctness, array_length(quiz_external_ids, 1)
	FROM %s WHERE shuffled_quiz_set_id = $1;`, e.TableName())

	err := database.Select(ctx, db, query, id).ScanFields(&totalCorrectness, &totalQuiz)
	if err != nil {
		return totalCorrectness, totalQuiz, err
	}
	return totalCorrectness, totalQuiz, nil
}

// GetShuffledQuizSetIDByOriginalQuizSetID returns shuffled_quiz_set_id by original_quiz_set_id
func (r *ShuffledQuizSetRepo) GetShuffledQuizSetIDsByOriginalQuizSetID(ctx context.Context, db database.QueryExecer, originalQuizSetID pgtype.Text) ([]string, error) {
	shuffledQuizSetIDs := []string{}
	e := entities.ShuffledQuizSet{}

	query := fmt.Sprintf(`SELECT shuffled_quiz_set_id FROM %s WHERE original_quiz_set_id = $1 AND deleted_at IS NULL`, e.TableName())

	rows, err := db.Query(ctx, query, originalQuizSetID)
	if err != nil {
		return nil, fmt.Errorf("ShuffledQuizSetRepo.GetShuffledQuizSetIDsByOriginalQuizSetID.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var shuffledQuizSetID string
		if err := rows.Scan(&shuffledQuizSetID); err != nil {
			return nil, fmt.Errorf("ShuffledQuizSetRepo.GetShuffledQuizSetIDsByOriginalQuizSetID.Scan: %w", err)
		}

		shuffledQuizSetIDs = append(shuffledQuizSetIDs, shuffledQuizSetID)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return shuffledQuizSetIDs, nil
}

// IsFinishedQuizTest check the quiz test has been finished
func (r *ShuffledQuizSetRepo) IsFinishedQuizTest(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Bool, error) {
	e := entities.ShuffledQuizSet{}

	isFinished := pgtype.Bool{}
	query := fmt.Sprintf(`SELECT (COUNT(DISTINCT (value->>'quiz_id')) = array_length(quiz_external_ids, 1)) AS is_finished
		FROM %s CROSS JOIN jsonb_array_elements(submission_history)
		WHERE shuffled_quiz_set_id = $1
		GROUP BY shuffled_quiz_set_id;`, e.TableName())

	err := database.Select(ctx, db, query, id).ScanFields(&isFinished)
	if err != nil {
		return isFinished, err
	}
	return isFinished, nil
}

// GetQuizIdx return quiz idx in quiz_external_ids
func (r *ShuffledQuizSetRepo) GetQuizIdx(ctx context.Context, db database.QueryExecer, id pgtype.Text, quizID pgtype.Text) (pgtype.Int4, error) {
	e := entities.ShuffledQuizSet{}
	idx := database.Int4(0)
	query := fmt.Sprintf(`SELECT ARRAY_POSITION(quiz_external_ids, $1) FROM %s WHERE shuffled_quiz_set_id=$2`, e.TableName())

	err := database.Select(ctx, db, query, quizID, id).ScanFields(&idx)
	if err != nil {
		return idx, err
	}
	return idx, nil
}

func (r *ShuffledQuizSetRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ShuffledQuizSet, error) {
	shuffledQuizSet := &entities.ShuffledQuizSet{}
	shuffledQuizSets := entities.ShuffledQuizSets{}
	values, _ := shuffledQuizSet.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND shuffled_quiz_set_id = ANY($1::_TEXT);`, strings.Join(values, ", "), shuffledQuizSet.TableName())

	if err := database.Select(ctx, db, stmt, ids).ScanAll(&shuffledQuizSets); err != nil {
		return nil, err
	}
	return shuffledQuizSets, nil
}

// GetExternalIDsFromSubmissionHistory get the external_ids which accepted from field `submission_history` with according shuffled_quiz_set_id
// how this query work?, if the data from is nil we will have empty array with `coalesce` handle it
// because the submission history have a lot record when student submit, maybe we will have duplicate data, so we use distinct.
// from another keyword, please use google search
func (r *ShuffledQuizSetRepo) GetExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text, isAccepted bool) (pgtype.TextArray, error) {
	var externalQuizIDs pgtype.TextArray
	stmtPlt := `SELECT coalesce(array_agg(DISTINCT(value ->>'quiz_id')),ARRAY[]::TEXT[]) FROM shuffled_quiz_sets CROSS JOIN jsonb_array_elements(submission_history) WHERE deleted_at IS NULL AND shuffled_quiz_set_id = $1`
	if isAccepted {
		stmtPlt += ` AND value ->>'is_accepted' = 'true';`
	} else {
		stmtPlt += ";"
	}
	if err := database.Select(ctx, db, stmtPlt, shuffleQuizSetID).ScanFields(&externalQuizIDs); err != nil {
		return pgtype.TextArray{}, err
	}
	return externalQuizIDs, nil
}

// ListExternalIDsFromSubmissionHistory
func (r *ShuffledQuizSetRepo) ListExternalIDsFromSubmissionHistory(ctx context.Context, db database.QueryExecer, shuffleQuizSetIDs pgtype.TextArray, isAccepted bool) (map[string][]string, error) {
	stmtPlt := `SELECT shuffled_quiz_set_id, coalesce(array_agg(DISTINCT(value ->>'quiz_id')),ARRAY[]::TEXT[]) FROM shuffled_quiz_sets CROSS JOIN jsonb_array_elements(submission_history) WHERE deleted_at IS NULL AND shuffled_quiz_set_id = ANY($1)`
	if isAccepted {
		stmtPlt += ` AND value ->>'is_accepted' = 'true'`
	}

	stmtPlt += " GROUP BY shuffled_quiz_set_id;"

	res := make(map[string][]string)
	rows, err := db.Query(ctx, stmtPlt, shuffleQuizSetIDs)
	if err != nil {
		return nil, fmt.Errorf("ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var externalQuizIDs pgtype.TextArray
		var shuffledQuizSetID pgtype.Text
		if err := rows.Scan(&shuffledQuizSetID, &externalQuizIDs); err != nil {
			return nil, fmt.Errorf("ShuffledQuizSetRepo.ListExternalIDsFromSubmissionHistory.Scan: %w", err)
		}

		for _, externalQuizID := range externalQuizIDs.Elements {
			res[shuffledQuizSetID.String] = append(res[shuffledQuizSetID.String], externalQuizID.String)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return res, nil
}

func (r *ShuffledQuizSetRepo) CalculateHighestSubmissionScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*CalculateHighestScoreResponse, error) {
	query := `
	SELECT sqs.study_plan_item_id,
       coalesce(max(coalesce(elss.graded_point,
                               (SELECT sum(POINT)
                                FROM get_submission_history() gsh
                                WHERE gsh.shuffled_quiz_set_id = sqs.shuffled_quiz_set_id))::float4 /
                      (SELECT sum(POINT)
                       FROM quizzes q
                       WHERE q.deleted_at IS NULL
                         AND q.external_id = ANY(array
                                                   (SELECT quiz_external_ids
                                                    FROM quiz_sets qs
                                                    WHERE qs.quiz_set_id = sqs.original_quiz_set_id
                                                      AND array_length(quiz_external_ids, 1) > 0)))) * 100, 0) AS percentage
	FROM shuffled_quiz_sets sqs
	LEFT JOIN exam_lo_submission els USING(shuffled_quiz_set_id)
	LEFT JOIN flash_card_submission fcs USING(shuffled_quiz_set_id)
	LEFT JOIN lo_submission ls USING(shuffled_quiz_set_id)
	LEFT JOIN get_exam_lo_returned_scores() elss ON elss.submission_id = els.submission_id
	WHERE sqs.study_plan_item_id = ANY($1::_TEXT)
	AND sqs.deleted_at IS NULL
	AND (fcs.is_submitted IS NULL
		OR fcs.is_submitted = TRUE)
	AND (ls.is_submitted IS NULL
		OR ls.is_submitted = TRUE)
	AND (elss.submission_id IS NULL
		OR elss.status = 'SUBMISSION_STATUS_RETURNED')
	GROUP BY sqs.study_plan_item_id`

	var res []*CalculateHighestScoreResponse
	rows, err := db.Query(ctx, query, &studyPlanItemIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Query: %w", err).Error())
	}
	defer rows.Close()

	for rows.Next() {
		var studyPlanItemID pgtype.Text
		var percentage float32
		if err := rows.Scan(&studyPlanItemID, &percentage); err != nil {
			return nil, fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Scan: %w", err)
		}

		res = append(res, &CalculateHighestScoreResponse{
			StudyPlanItemID: studyPlanItemID,
			Percentage:      database.Float4(percentage),
		})
	}
	return res, nil
}

// GetByExternalIDs Retrieve retrieves a multiple Quiz by external id
func (r *QuizRepo) GetByExternalIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, loID pgtype.Text) (entities.Quizzes, error) {
	e := &entities.Quiz{}
	loe := &entities.LearningObjective{}

	fields, _ := e.FieldMap()
	for i := range fields {
		fields[i] = "q." + fields[i]
	}
	stmt := fmt.Sprintf(`SELECT %s
			FROM %s q INNER JOIN %s lo ON q.school_id = lo.school_id
			INNER JOIN UNNEST($1::TEXT[]) WITH ORDINALITY AS search_quiz_external_ids(id, id_order)
			ON q.external_id = search_quiz_external_ids.id
			WHERE q.deleted_at IS NULL
			AND lo.lo_id = $2
			ORDER BY search_quiz_external_ids.id_order;`,
		strings.Join(fields, ","), e.TableName(), loe.TableName())

	quizzes := entities.Quizzes{}
	if err := database.Select(ctx, db, stmt, &ids, &loID).ScanAll(&quizzes); err != nil {
		return nil, err
	}

	return quizzes, nil
}

func (r *ShuffledQuizSetRepo) GenerateExamLOSubmission(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (*entities.ExamLOSubmission, error) {
	ctx, span := interceptors.StartSpan(ctx, "ShuffledQuizSetRepo.GenerateExamLOSubmission")
	defer span.End()

	var results entities.ExamLOSubmission

	stmt := `
    SELECT generate_ulid() AS submission_id,
           SQ.student_id,
           SQ.study_plan_id,
           SQ.learning_material_id,
           SQ.shuffled_quiz_set_id,
           NULL AS status,
           NULL AS result,
           NULL AS teacher_feedback,
           NULL AS teacher_id,
           NULL AS marked_at,
           NULL AS removed_at,
           COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(SQ.quiz_external_ids)), 0)::INT AS total_point,
           SQ.updated_at as created_at, -- elsa.created_at == sqs.updated_at
           SQ.updated_at, -- els.created_at & els.updated_at is the same.
           SQ.deleted_at
      FROM shuffled_quiz_sets SQ
     WHERE shuffled_quiz_set_id = $1::TEXT;
	`

	if err := database.Select(ctx, db, stmt, shuffleQuizSetID).ScanOne(&results); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return &results, nil
}

func (r *ShuffledQuizSetRepo) GetExternalIDs(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (pgtype.TextArray, error) {
	ctx, span := interceptors.StartSpan(ctx, "ShuffledQuizSetRepo.GetExternalIDs")
	defer span.End()

	var result pgtype.TextArray

	stmt := `SELECT quiz_external_ids FROM shuffled_quiz_sets WHERE shuffled_quiz_set_id = $1::TEXT;`

	if err := database.Select(ctx, db, stmt, shuffleQuizSetID).ScanFields(&result); err != nil {
		return pgtype.TextArray{}, fmt.Errorf("database.Select: %w", err)
	}

	return result, nil
}

func (r *QuizRepo) GetQuizByExternalID(ctx context.Context, db database.QueryExecer, externalID pgtype.Text) (*entities.Quiz, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuizRepo.GetQuizByExternalID")
	defer span.End()

	var result entities.Quiz
	fields, _ := result.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE deleted_at IS NULL AND external_id = $1::TEXT;`, strings.Join(fields, ","), result.TableName())
	if err := database.Select(ctx, db, stmt, externalID).ScanOne(&result); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return &result, nil
}

func (r *ShuffledQuizSetRepo) GetCorrectnessInfo(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text, externalID pgtype.Text) (*entities.CorrectnessInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "ShuffledQuizSetRepo.GetCorrectnessInfo")
	defer span.End()

	var result entities.CorrectnessInfo

	stmt := `
    SELECT random_seed,
           coalesce(array_position(quiz_external_ids, $2::TEXT), 0)::INT AS quiz_index,
           (SELECT count(DISTINCT (value ->> 'quiz_id'))
              FROM public.shuffled_quiz_sets X
                  CROSS JOIN jsonb_array_elements(X.submission_history)
             WHERE X.shuffled_quiz_set_id = SQ.shuffled_quiz_set_id)::INT AS total_submission_history,
           cardinality(quiz_external_ids)::INT AS total_quiz_external_ids,
           SQ.original_shuffle_quiz_set_id,
           SQ.student_id,
           SQ.total_correctness::INT,
           (SELECT lo_id FROM quiz_sets QS where QS.deleted_at IS NULL AND QS.quiz_set_id = SQ.original_quiz_set_id) AS lo_id
      FROM shuffled_quiz_sets SQ
     WHERE SQ.deleted_at IS NULL
       AND SQ.shuffled_quiz_set_id = $1::TEXT;
	`

	if err := database.Select(ctx, db, stmt, shuffleQuizSetID, externalID).ScanOne(&result); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return &result, nil
}

func (r *ShuffledQuizSetRepo) UpdateTotalCorrectnessAndSubmissionHistory(ctx context.Context, db database.QueryExecer, e *entities.ShuffledQuizSet) error {
	ctx, span := interceptors.StartSpan(ctx, "ShuffledQuizSetRepo.UpdateTotalCorrectnessAndSubmissionHistory")
	defer span.End()

	stmt := fmt.Sprintf(`
    UPDATE %s
       SET total_correctness = $1,
           submission_history = submission_history || $2,
           updated_at = $3
     WHERE shuffled_quiz_set_id = $4;
	`, e.TableName())

	cmdTag, err := db.Exec(ctx, stmt, e.TotalCorrectness, e.SubmissionHistory, e.UpdatedAt, e.ID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no row affected")
	}

	return nil
}

func (r *ShuffledQuizSetRepo) UpsertLOSubmission(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (*entities.LOSubmissionAnswerKey, error) {
	ctx, span := interceptors.StartSpan(ctx, "ShuffledQuizSetRepo.UpsertLOSubmission")
	defer span.End()

	result := &entities.LOSubmissionAnswerKey{}
	fields, values := result.FieldMap()

	stmt := fmt.Sprintf(`
    INSERT INTO lo_submission (
        submission_id,
        student_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        created_at,
        updated_at,
        deleted_at,
        total_point
    )
    SELECT generate_ulid() AS submission_id,
           SQ.student_id,
           SQ.study_plan_id,
           SQ.learning_material_id,
           SQ.shuffled_quiz_set_id,
           SQ.created_at,
           SQ.updated_at,
           SQ.deleted_at,
           COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(SQ.quiz_external_ids)), 0)::INT AS total_point
    FROM shuffled_quiz_sets SQ
    WHERE shuffled_quiz_set_id = $1::TEXT
    ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO UPDATE SET
        updated_at = EXCLUDED.updated_at,
        total_point = EXCLUDED.total_point
    RETURNING %s;
	`, strings.Join(fields, ","))

	err := db.QueryRow(ctx, stmt, shuffleQuizSetID).Scan(values...)
	if err != nil {
		return result, fmt.Errorf("db.QueryRow: %w", err)
	}
	return result, nil
}

func (r *ShuffledQuizSetRepo) UpsertFlashCardSubmission(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (*entities.FlashCardSubmissionAnswerKey, error) {
	ctx, span := interceptors.StartSpan(ctx, "ShuffledQuizSetRepo.UpsertFlashCardSubmission")
	defer span.End()

	result := &entities.FlashCardSubmissionAnswerKey{}
	fields, values := result.FieldMap()

	stmt := fmt.Sprintf(`
    INSERT INTO flash_card_submission (
        submission_id,
        student_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        created_at,
        updated_at,
        deleted_at,
        total_point
    )
    SELECT generate_ulid() AS submission_id,
           SQ.student_id,
           SQ.study_plan_id,
           SQ.learning_material_id,
           SQ.shuffled_quiz_set_id,
           SQ.created_at,
           SQ.updated_at,
           SQ.deleted_at,
           COALESCE((SELECT SUM(point) FROM public.quizzes q WHERE q.deleted_at IS NULL AND q.external_id = ANY(SQ.quiz_external_ids)), 0)::INT AS total_point
    FROM shuffled_quiz_sets SQ
    WHERE shuffled_quiz_set_id = $1::TEXT
    ON CONFLICT ON CONSTRAINT flash_card_submission_shuffled_quiz_set_id_un DO UPDATE SET
        updated_at = EXCLUDED.updated_at,
        total_point = EXCLUDED.total_point
    RETURNING %s;
	`, strings.Join(fields, ","))

	err := db.QueryRow(ctx, stmt, shuffleQuizSetID).Scan(values...)
	if err != nil {
		return result, fmt.Errorf("db.QueryRow: %w", err)
	}
	return result, nil
}

func (r *ShuffledQuizSetRepo) GetRelatedLearningMaterial(ctx context.Context, db database.QueryExecer, shuffleQuizSetID pgtype.Text) (*entities.LearningMaterial, error) {
	sqs := entities.ShuffledQuizSet{}

	lm := &entities.LearningMaterial{}
	fields, values := lm.FieldMap()

	stmt := fmt.Sprintf(`
	SELECT
		lm.%s
	FROM
		%s sqs
		JOIN %s lm USING(learning_material_id)
	WHERE
		sqs.shuffled_quiz_set_id = $1::TEXT;
	`, strings.Join(fields, ", lm."), sqs.TableName(), lm.TableName())

	if err := db.QueryRow(ctx, stmt, shuffleQuizSetID).Scan(values...); err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return lm, nil
}

func (r *QuizRepo) GetTagNames(ctx context.Context, db database.QueryExecer, externalIDs pgtype.TextArray) (map[string][]string, error) {
	tag := entities.QuestionTag{}
	stmt := fmt.Sprintf(`
	SELECT
		q.external_id,
		array_agg(qt.name)
	FROM
		quizzes q
	JOIN %s qt ON
		qt.question_tag_id = ANY(q.question_tag_ids)
	WHERE
		external_id = ANY($1)
		AND qt.deleted_at IS NULL
		AND q.deleted_at IS NULL
	GROUP BY q.external_id
	`, tag.TableName())
	rows, err := db.Query(ctx, stmt, externalIDs)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	defer rows.Close()
	tagMap := make(map[string][]string)
	for rows.Next() {
		var (
			externalID string
			tagNames   []string
		)
		if err := rows.Scan(&externalID, &tagNames); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		tagMap[externalID] = tagNames
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return tagMap, nil
}
