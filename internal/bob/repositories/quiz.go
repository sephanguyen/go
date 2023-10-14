package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// QuizRepo works with quizs
type QuizRepo struct{}

const quizRepoSearchStmt = `SELECT %s FROM quizzes
WHERE deleted_at IS NULL
AND external_id = ANY($1)
AND ($2::text IS NULL OR status = $2)
ORDER BY quiz_id DESC
LIMIT %d`

const quizRepoDeleteByExternalIDStmt = `UPDATE quizzes
SET deleted_at = NOW(), status = 'QUIZ_STATUS_DELETED'
WHERE external_id = $1 AND school_id = $2`

const quizRepoRetrieveStmt = `SELECT %s FROM quizzes
WHERE quiz_id = $1`

// Create creates Quiz
func (r *QuizRepo) Create(ctx context.Context, db database.QueryExecer, quiz *entities_bob.Quiz) error {
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

func (r *QuizRepo) Upsert(ctx context.Context, db database.QueryExecer, data []*entities_bob.Quiz) ([]*entities_bob.Quiz, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuizRepo.Upsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, t *entities_bob.Quiz) {
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
				created_by = $12,
				approved_by = $13,
				status = $14,
				updated_at = $15,
				created_at = $16,
				deleted_at = $17
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

	resp := make([]*entities_bob.Quiz, 0)
	for i := 0; i < b.Len(); i++ {
		quiz := new(entities_bob.Quiz)
		_, values := quiz.FieldMap()
		if err := result.QueryRow().Scan(values...); err != nil {
			return nil, fmt.Errorf("batchResults.QueryRow: %w", err)
		}
		resp = append(resp, quiz)
	}

	return resp, nil
}

// QuizFilter can be use with Search
type QuizFilter struct {
	ExternalIDs pgtype.TextArray
	Status      pgtype.Text
	Limit       uint
}

// Search returns by user ID
func (r *QuizRepo) Search(ctx context.Context, db database.QueryExecer, filter QuizFilter) (entities_bob.Quizzes, error) {
	e := &entities_bob.Quiz{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(quizRepoSearchStmt, strings.Join(fields, ","), filter.Limit)

	results := make(entities_bob.Quizzes, 0, filter.Limit)
	err := database.Select(ctx, db, stmt,
		&filter.ExternalIDs,
		&filter.Status).ScanAll(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetByExternalID retrieves a single Quiz by external id
func (r *QuizRepo) GetByExternalID(ctx context.Context, db database.QueryExecer, id pgtype.Text, schoolID pgtype.Int4) (*entities_bob.Quiz, error) {
	e := &entities_bob.Quiz{}

	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
			FROM %s
			WHERE deleted_at is NULL
			AND external_id = $1
			AND school_id = $2`,
		strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, stmt, &id, &schoolID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

// GetOptions get options of a quiz by its ID
func (r *QuizRepo) GetOptions(ctx context.Context, db database.QueryExecer, quizID pgtype.Text, loID pgtype.Text) ([]*entities_bob.QuizOption, error) {
	e := &entities_bob.Quiz{}
	loe := &entities.LearningObjective{}
	optionsJSONB := database.JSONB("")
	stmt := fmt.Sprintf(`SELECT options
			FROM %s LEFT JOIN %s ON %s.school_id = %s.school_id
			WHERE %s.deleted_at is NULL
			AND external_id = $1
			AND %s.lo_id=$2;`,
		e.TableName(), loe.TableName(), e.TableName(), loe.TableName(), e.TableName(), loe.TableName(),
	)

	if err := db.QueryRow(ctx, stmt, quizID, loID).Scan(&optionsJSONB); err != nil {
		return nil, err
	}

	options := make([]*entities_bob.QuizOption, 0)
	err := optionsJSONB.AssignTo(&options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

// Retrieve retrieves a single Quiz
func (r *QuizRepo) Retrieve(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities_bob.Quiz, error) {
	e := &entities_bob.Quiz{}
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

// GetByExternalIDs Retrieve retrieves a multiple Quiz by external id
func (r *QuizRepo) GetByExternalIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, loID pgtype.Text) (entities_bob.Quizzes, error) {
	e := &entities_bob.Quiz{}
	loe := &entities.LearningObjective{}

	fields, _ := e.FieldMap()
	for i := range fields {
		fields[i] = "q." + fields[i]
	}
	stmt := fmt.Sprintf(`SELECT %s
			FROM %s q INNER JOIN %s lo ON q.school_id = lo.school_id
			INNER JOIN unnest($1::text[]) WITH ORDINALITY AS search_quiz_external_ids(id, id_order)
			ON q.external_id = search_quiz_external_ids.id
			WHERE q.deleted_at is NULL
			AND lo.lo_id = $2
			ORDER BY search_quiz_external_ids.id_order;`,
		strings.Join(fields, ","), e.TableName(), loe.TableName())

	quizzes := entities_bob.Quizzes{}
	if err := database.Select(ctx, db, stmt, &ids, &loID).ScanAll(&quizzes); err != nil {
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
type QuizSetRepo struct{}

const quizSetRepoSearchStmt = `SELECT %s FROM quiz_sets
WHERE deleted_at IS NULL
AND ($1::text[] IS NULL OR lo_id = ANY ($1))
AND ($2::text IS NULL OR status = $2)
ORDER BY quiz_set_id DESC
LIMIT %d`

const quizSetRepoDeleteStmt = `UPDATE quiz_sets
SET deleted_at = NOW(), status = 'QUIZSET_STATUS_DELETED', updated_at = NOW()
WHERE quiz_set_id = $1`

// Search searches quiz_sets match the filter
func (r *QuizSetRepo) Search(ctx context.Context, db database.QueryExecer, filter QuizSetFilter) (entities_bob.QuizSets, error) {
	if len(filter.ObjectiveIDs.Elements) == 0 {
		return nil, nil
	}

	e := &entities_bob.QuizSet{}
	stmt := fmt.Sprintf(quizSetRepoSearchStmt, strings.Join(database.GetFieldNames(e), ","), filter.Limit)

	results := make(entities_bob.QuizSets, 0, filter.Limit)
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
func (r *QuizSetRepo) Create(ctx context.Context, db database.QueryExecer, e *entities_bob.QuizSet) error {
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
func (r *QuizSetRepo) GetQuizSetByLoID(ctx context.Context, db database.QueryExecer, loID pgtype.Text) (*entities_bob.QuizSet, error) {
	quizSet := &entities_bob.QuizSet{}

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
func (r *QuizSetRepo) GetQuizSetsOfLOContainQuiz(ctx context.Context, db database.QueryExecer, loID pgtype.Text, quizID pgtype.Text) (entities_bob.QuizSets, error) {
	quizSets := entities_bob.QuizSets{}
	quizSet := entities_bob.QuizSet{}
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
func (r *QuizSetRepo) GetQuizSetsContainQuiz(ctx context.Context, db database.QueryExecer, quizID pgtype.Text) (entities_bob.QuizSets, error) {
	quizSets := entities_bob.QuizSets{}
	quizSet := entities_bob.QuizSet{}
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

// GetTotalQuiz return total quiz array of lo_ids
func (r *QuizSetRepo) GetTotalQuiz(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error) {
	e := &entities_bob.QuizSet{}

	res := make(map[string]int32)

	stmt := fmt.Sprintf(`SELECT lo_id, array_length(quiz_external_ids, 1)
	FROM %s
	WHERE deleted_at IS NULL AND lo_id = ANY($1)`, e.TableName())

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

	for _, id := range loIDs.Elements {
		if _, ok := res[id.String]; !ok {
			res[id.String] = 0
		}
	}

	return res, nil
}

// GetQuizExternalIDs returns quiz external ids of a quiz set
func (r *QuizSetRepo) GetQuizExternalIDs(ctx context.Context, db database.QueryExecer, loID pgtype.Text, limit pgtype.Int8, offset pgtype.Int8) ([]string, error) {
	quizSet := &entities_bob.QuizSet{}

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

func (r *QuizSetRepo) CountQuizOnLO(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionSetRepo.CountQuizOnLO")
	defer span.End()

	e := &entities_bob.QuizSet{}

	query := fmt.Sprintf("SELECT lo_id, array_length(quiz_external_ids, 1) FROM %s WHERE lo_id = ANY($1) AND status = $2", e.TableName())

	rows, err := db.Query(ctx, query, &loIDs, "QUIZSET_STATUS_APPROVED")
	if err != nil {
		return nil, errors.Wrap(err, "repo.DB.QueryEx")
	}
	defer rows.Close()

	ss := make(map[string]int32)

	defer rows.Close()
	for rows.Next() {
		var loID pgtype.Text
		var count pgtype.Int4
		if err := rows.Scan(&loID, &count); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		// only choose an quiz_external_ids not null
		if count.Status == pgtype.Present {
			ss[loID.String] = count.Int
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return ss, nil
}

// ShuffledQuizSetRepo quizzes in the quiz set will be shuffled
type ShuffledQuizSetRepo struct{}

// Get return shuffledQuizSet by id
func (r *ShuffledQuizSetRepo) Get(ctx context.Context, db database.QueryExecer, id pgtype.Text, from pgtype.Int8, to pgtype.Int8) (*entities_bob.ShuffledQuizSet, error) {
	shuffledQuizSet := &entities_bob.ShuffledQuizSet{}

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
	shuffledQuizSet := &entities_bob.ShuffledQuizSet{}

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
func (r *ShuffledQuizSetRepo) GetByStudyPlanItems(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) (entities_bob.ShuffledQuizSets, error) {
	ents := entities_bob.ShuffledQuizSets{}
	e := entities_bob.ShuffledQuizSet{}

	fieldNames := database.GetFieldNames(&e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_plan_item_id = ANY($1)", strings.Join(fieldNames, ","), e.TableName())

	err := database.Select(ctx, db, query, studyPlanItemIDs).ScanAll(&ents)
	if err != nil {
		return ents, err
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
	e := entities_bob.ShuffledQuizSet{}

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
	e := entities_bob.ShuffledQuizSet{}

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
	shuffledQuizSet := entities_bob.ShuffledQuizSet{}
	quizSet := entities_bob.QuizSet{}

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
	e := entities_bob.ShuffledQuizSet{}

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

// IsFinishedQuizTest check the quiz test has been finished
func (r *ShuffledQuizSetRepo) IsFinishedQuizTest(ctx context.Context, db database.QueryExecer, id pgtype.Text) (pgtype.Bool, error) {
	e := entities_bob.ShuffledQuizSet{}

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
	e := entities_bob.ShuffledQuizSet{}
	idx := database.Int4(0)
	query := fmt.Sprintf(`SELECT ARRAY_POSITION(quiz_external_ids, $1) FROM %s WHERE shuffled_quiz_set_id=$2`, e.TableName())

	err := database.Select(ctx, db, query, quizID, id).ScanFields(&idx)
	if err != nil {
		return idx, err
	}
	return idx, nil
}

func (r *ShuffledQuizSetRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities_bob.ShuffledQuizSet, error) {
	shuffledQuizSet := &entities_bob.ShuffledQuizSet{}
	shuffledQuizSets := entities_bob.ShuffledQuizSets{}
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

type CalculateHighestScoreResponse struct {
	StudyPlanItemID pgtype.Text
	Percentage      pgtype.Float4
}

func (r *ShuffledQuizSetRepo) CalculateHigestSubmissionScore(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*CalculateHighestScoreResponse, error) {
	query := `
	SELECT
	sqs.study_plan_item_id,
	max( sqs.total_correctness::float4 / array_length(qs.quiz_external_ids::text[], 1)::float4 * 100)::numeric as percentage
	FROM shuffled_quiz_sets sqs
	JOIN quiz_sets qs ON sqs.original_quiz_set_id = qs.quiz_set_id
	JOIN learning_objectives lo ON lo.lo_id = qs.lo_id
	WHERE  sqs.study_plan_item_id = ANY($1::text[])
			AND sqs.deleted_at IS NULL
			AND qs.deleted_at IS NULL
			AND lo.deleted_at IS NULL
			AND array_length(qs.quiz_external_ids::text[], 1) is not null
	GROUP BY study_plan_item_id	
	`

	var res []*CalculateHighestScoreResponse
	rows, err := db.Query(ctx, query, &studyPlanItemIDs)
	if err != nil {
		return nil, fmt.Errorf("ShuffledQuizSetRepo.CalculateHigestSubmissionScore.Query: %w", err)
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
