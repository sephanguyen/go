package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// QuestionRepo stores
type QuestionRepo struct{}

func insertQuestionBatch(ctx context.Context, db database.QueryExecer, questions []entities.Question) ([]*entities.Question, error) {
	b := &pgx.Batch{}
	var updateQuestions []*entities.Question

	queueFn := func(b *pgx.Batch, q *entities.Question) {
		fieldNames := []string{"question_id", "master_question_id", "country", "question", "answers", "explanation", "difficulty_level", "updated_at", "created_at", "question_rendered", "answers_rendered", "explanation_rendered", "is_waiting_for_render", "explanation_wrong_answer", "explanation_wrong_answer_rendered"}
		placeHolders := "$1, $2, $3, $4, $5::text[], $6, $7, $8, $9, $10, $11, $12, $13, $14, $15"

		query := fmt.Sprintf("INSERT INTO %s as A (%s) "+
			"VALUES (%s) ON CONFLICT ON CONSTRAINT questions_pk "+
			"DO UPDATE SET master_question_id = $2, country = $3, question = $4, answers = $5::text[], explanation = $6, difficulty_level = $7, updated_at = $8, is_waiting_for_render = $13, explanation_wrong_answer = $14 "+
			"WHERE A.master_question_id <> $2 OR A.country <> $3 OR A.question <> $4 OR A.answers <> $5::text[] "+
			"OR A.explanation <> $6 OR A.difficulty_level <> $7 OR A.explanation_wrong_answer <> $14 "+
			"RETURNING A.question_id", q.TableName(), strings.Join(fieldNames, ","), placeHolders)

		b.Queue(query, database.GetScanFields(q, fieldNames)...)
	}

	for _, q := range questions {
		question := q
		_ = question.IsWaitingForRender.Set(true)
		queueFn(b, &question)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < b.Len(); i++ {
		row := batchResults.QueryRow()

		updateQuestion := new(entities.Question)
		if err := row.Scan(&updateQuestion.QuestionID); err != pgx.ErrNoRows && err != nil {
			return updateQuestions, err
		}

		if updateQuestion.QuestionID.String != "" {
			updateQuestions = append(updateQuestions, updateQuestion)
		}
	}

	return updateQuestions, nil
}

func insertQuestionTagLoBatch(ctx context.Context, db database.QueryExecer, questions []entities.Question, questionTagLo []entities.QuestionTagLo) error {
	insertTagLoFn := func(b *pgx.Batch, tag *entities.QuestionTagLo) {
		fieldNames, _ := tag.FieldMap()
		placeHolders := "$1, $2, $3"

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT questions_tagged_learning_objectives_pk DO NOTHING", tag.TableName(), strings.Join(fieldNames, ","), placeHolders)

		sc := database.GetScanFields(tag, fieldNames)
		b.Queue(query, sc...)
	}

	if len(questionTagLo) == 0 {
		return nil
	}

	b := &pgx.Batch{}
	for _, q := range questions {
		displayOrder := 1
		for _, tag := range questionTagLo {
			if q.QuestionID == tag.QuestionID {
				tagLo := tag
				tagLo.DisplayOrder.Set(displayOrder)
				insertTagLoFn(b, &tagLo)
				displayOrder++
			}
		}
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(questionTagLo); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("Tag LO not inserted")
		}
	}
	return nil
}

// CreateAll create list of question into DB
func (repo *QuestionRepo) CreateAll(ctx context.Context, db database.QueryExecer, questions []entities.Question, questionTagLo []entities.QuestionTagLo) ([]*entities.Question, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionRepo.CreateAll")
	defer span.End()

	questionIds := make([]string, 0, len(questions))
	for _, question := range questions {
		questionIds = append(questionIds, question.QuestionID.String)
	}

	e := new(entities.QuestionTagLo)
	var Ids pgtype.TextArray
	Ids.Set(questionIds)

	updateQuestions, err := insertQuestionBatch(ctx, db, questions)
	if err != nil {
		return nil, errors.Wrap(err, "error insert question")
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE question_id=ANY($1)`, e.TableName())
	_, err = db.Exec(ctx, query, &Ids)
	if err != nil {
		return nil, fmt.Errorf("db.Exec: %w", err)
	}

	if err := insertQuestionTagLoBatch(ctx, db, questions, questionTagLo); err != nil {
		return nil, errors.Wrap(err, "error insert tag lo")
	}

	return updateQuestions, nil
}

// ExistMasterQuestion is use to check if masterQuestionId is valid
func (repo *QuestionRepo) ExistMasterQuestion(ctx context.Context, db database.QueryExecer, masterQuestionId string) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionRepo.ExistMasterQuestion")
	defer span.End()

	query := `SELECT COUNT (*) FROM questions WHERE question_id = $1`
	var questionID pgtype.Text
	questionID.Set(masterQuestionId)
	row := db.QueryRow(ctx, query, &questionID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	if count < 1 {
		return false, nil
	}
	return true, nil
}

// RetrieveQuestionsFromLoId retrieves list of question from one Lo Id
func (repo *QuestionRepo) RetrieveQuestionsFromLoId(ctx context.Context, db database.QueryExecer, loId pgtype.Text) ([]*entities.Question, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionRepo.RetrieveQuestionsFromLoId")
	defer span.End()

	q := &entities.Question{}
	fields := database.GetFieldNames(q)
	query := `SELECT %s FROM %s a INNER JOIN (SELECT %s FROM questions_tagged_learning_objectives WHERE lo_id =$1) b USING(%s)`
	smt := fmt.Sprintf(query, strings.Join(fields, ","), q.TableName(), "question_id", "question_id")
	rows, err := db.Query(ctx, smt, &loId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questionList []*entities.Question
	for rows.Next() {
		question := new(entities.Question)
		if err := rows.Scan(database.GetScanFields(question, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		questionList = append(questionList, question)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return questionList, nil
}

type QuestionPagination struct {
	Questions []*entities.Question
	Total     pgtype.Int8
}

// RetrieveQuizSetsFromLoId return quizset belong to a learning objective
func (repo *QuestionRepo) RetrieveQuizSetsFromLoId(ctx context.Context, db database.QueryExecer, loId pgtype.Text, topicType pgtype.Text, limit, page int) (*QuestionPagination, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionRepo.RetrieveQuizSetsFromLoId")
	defer span.End()

	q := &entities.Question{}
	tpType := pb.TopicType(pb.TopicType_value[topicType.String])
	questionFieldsName := []string{
		"question_id", "master_question_id", "country", "question", "answers", "difficulty_level", "updated_at", "created_at",
		"question_rendered", "answers_rendered", "is_waiting_for_render",
		"question_url", "answers_url", "rendering_question",
	}

	args := []interface{}{&loId}
	query := `SELECT %s, COUNT(*) OVER() AS total FROM %s a INNER JOIN (SELECT %s FROM quizsets WHERE lo_id = $1) b USING(%s) ORDER BY display_order`

	// only use limit and offset when page > 0 to ensure backward-compatible
	if page > 0 {
		if limit == 0 {
			limit = 5
		}
		args = append(args, limit, limit*(page-1))
		query += " LIMIT $2 OFFSET $3"
	}
	quizSetsFieldName := []string{"lo_id", "question_id", "display_order"}

	if tpType != pb.TOPIC_TYPE_EXAM {
		explainFields := []string{
			"explanation", "explanation_wrong_answer",
			"explanation_rendered", "explanation_wrong_answer_rendered", "explanation_url", "explanation_wrong_answer_url",
		}
		questionFieldsName = append(questionFieldsName, explainFields...)
	}
	smt := fmt.Sprintf(query, strings.Join(questionFieldsName, ","), q.TableName(), strings.Join(quizSetsFieldName, ","), "question_id")
	rows, err := db.Query(ctx, smt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		questions []*entities.Question
		total     pgtype.Int8
	)
	for rows.Next() {
		question := new(entities.Question)
		fields := append(database.GetScanFields(question, questionFieldsName), &total)
		if err := rows.Scan(fields...); err != nil {
			return nil, fmt.Errorf("rows.Scan %w", err)
		}
		questions = append(questions, question)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Scan %w", err)
	}
	return &QuestionPagination{questions, total}, nil
}

func (repo *QuestionRepo) RetrieveQuestionTagLo(ctx context.Context, db database.QueryExecer, questionIds pgtype.TextArray) (map[string][]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionRepo.RetrieveQuizSetsFromLoId")
	defer span.End()

	q := &entities.QuestionTagLo{}
	fields := database.GetFieldNames(q)

	query := `SELECT %s FROM %s WHERE question_id=ANY($1) ORDER BY display_order ASC`
	smt := fmt.Sprintf(query, strings.Join(fields, ","), q.TableName())
	rows, err := db.Query(ctx, smt, &questionIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taggedLoMap := make(map[string][]string)
	for rows.Next() {
		taggedLo := new(entities.QuestionTagLo)
		if err := rows.Scan(database.GetScanFields(taggedLo, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		questionId := taggedLo.QuestionID.String
		loId := taggedLo.LoID.String
		taggedLoMap[questionId] = append(taggedLoMap[questionId], loId)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return taggedLoMap, nil
}

func (repo *QuestionRepo) RetrieveQuiz(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Question, error) {
	ctx, span := interceptors.StartSpan(ctx, "QuestionRepo.RetrieveQuiz")
	defer span.End()

	e := &entities.Question{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE question_id = $1", strings.Join(fields, ","), e.TableName())
	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryRowEx")
	}

	return e, nil
}
