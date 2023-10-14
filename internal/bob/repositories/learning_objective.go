package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LearningObjectiveRepo struct{}

func (r *LearningObjectiveRepo) Create(ctx context.Context, db database.QueryExecer, m *entities.LearningObjective) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.Create")
	defer span.End()

	now := time.Now()
	m.UpdatedAt.Set(now)
	m.CreatedAt.Set(now)
	m.DeletedAt.Set(nil)

	cmdTag, err := database.Insert(ctx, m, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new " + m.TableName())
	}

	return nil
}

func (r *LearningObjectiveRepo) BulkImport(ctx context.Context, db database.QueryExecer, learningObjectives []*entities.LearningObjective) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.BulkImport")
	defer span.End()

	queueFn := func(b *pgx.Batch, p *entities.LearningObjective) {
		fieldNames := database.GetFieldNames(p)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := `INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT learning_objectives_pk
		DO UPDATE SET name = $2, country = $3, grade = $4, subject = $5, topic_id = $6, master_lo_id = $7, video_script = $9, prerequisites = $10, video = $11, study_guide = $12, school_id = $13,  updated_at = $15, type = $17`
		b.Queue(fmt.Sprintf(query, p.TableName(), strings.Join(fieldNames, ","), placeHolders), database.GetScanFields(p, fieldNames)...)
	}

	var d pgtype.Timestamptz
	d.Set(time.Now())

	b := &pgx.Batch{}
	for _, lo := range learningObjectives {
		lo.CreatedAt = d
		lo.UpdatedAt = d
		queueFn(b, lo)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(learningObjectives); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("learning objectives not inserted")
		}
	}

	return nil
}

func (r *LearningObjectiveRepo) RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.RetrieveByTopicIDs")
	defer span.End()

	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	query := fmt.Sprintf("SELECT DISTINCT lo.%s FROM %s lo LEFT JOIN %s t ON t.topic_id = lo.topic_id WHERE lo.topic_id = ANY($1) AND lo.deleted_at IS NULL AND t.deleted_at IS NULL", strings.Join(fields, ", lo."), lo.TableName(), (&entities.Topic{}).TableName())
	rows, err := db.Query(ctx, query, &topicIds)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*entities.LearningObjective
	for rows.Next() {
		p := new(entities.LearningObjective)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *LearningObjectiveRepo) FindInQuestionTagLo(ctx context.Context, db database.QueryExecer, questionIds pgtype.TextArray) (mapQuestionIdLo map[pgtype.Text]*entities.LearningObjective, err error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.findInQuestionTagLo")
	defer span.End()

	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	query := fmt.Sprintf("SELECT lo.%s, qt.question_id "+
		"FROM %s AS lo JOIN questions_tagged_learning_objectives AS qt ON lo.lo_id = qt.lo_id "+
		"WHERE qt.question_id = ANY($1) and lo.deleted_at IS NULL", strings.Join(fields, ", lo."), lo.TableName())
	rows, err := db.Query(ctx, query, &questionIds)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	mapQuestionIdLo = make(map[pgtype.Text]*entities.LearningObjective)
	for rows.Next() {
		p := new(entities.LearningObjective)
		var questionId pgtype.Text
		if err := rows.Scan(append(database.GetScanFields(p, fields), &questionId)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		mapQuestionIdLo[questionId] = p
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return mapQuestionIdLo, nil
}

func (r *LearningObjectiveRepo) FindInQuizSet(ctx context.Context, db database.QueryExecer, questionIds pgtype.TextArray) (mapQuestionIdLo map[pgtype.Text]*entities.LearningObjective, err error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.findInQuizSet")
	defer span.End()

	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	query := fmt.Sprintf(`SELECT lo.%s, qt.question_id
		FROM %s AS lo JOIN quizsets AS qt ON lo.lo_id = qt.lo_id
		WHERE qt.question_id = ANY($1) AND lo.deleted_at IS NULL
		LIMIT 1`, strings.Join(fields, ", lo."), lo.TableName())
	rows, err := db.Query(ctx, query, &questionIds)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	mapQuestionIdLo = make(map[pgtype.Text]*entities.LearningObjective)
	for rows.Next() {
		p := new(entities.LearningObjective)
		var questionId pgtype.Text
		if err := rows.Scan(append(database.GetScanFields(p, fields), &questionId)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		mapQuestionIdLo[questionId] = p
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return mapQuestionIdLo, nil
}

func (r *LearningObjectiveRepo) SuggestByLOName(ctx context.Context, db database.QueryExecer, loName string) ([]*entities.LearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.RetrieveByTopicIDs")
	defer span.End()

	searchName := fmt.Sprintf("%s%s%s", "%", loName, "%")
	lo := &entities.LearningObjective{}
	fields := []string{"lo_id", "name"}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE name ILIKE $1 AND deleted_at IS NULL LIMIT %d", strings.Join(fields, ","), lo.TableName(), 10)
	rows, err := db.Query(ctx, query, &searchName)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*entities.LearningObjective
	for rows.Next() {
		p := new(entities.LearningObjective)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return pp, nil
}

func (r *LearningObjectiveRepo) RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.RetrieveByTopicIDs")
	defer span.End()

	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lo_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), lo.TableName())
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var pp []*entities.LearningObjective
	for rows.Next() {
		p := new(entities.LearningObjective)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

func (r *LearningObjectiveRepo) DuplicateLearningObjectives(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray, newTopicIDs pgtype.TextArray) ([]*entities.CopiedLearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.DuplicateLearningObjectives")
	defer span.End()

	var now pgtype.Timestamptz
	now.Set(time.Now().UTC())

	queue := func(b *pgx.Batch, topicID pgtype.Text, newTopicID pgtype.Text) {
		e := &entities.LearningObjective{}
		topicFieldNames := database.GetFieldNames(e)
		selectFields := golibs.Replace(
			topicFieldNames,
			[]string{"lo_id", "copied_from", "topic_id", "created_at", "updated_at"},
			[]string{"uuid_generate_v4()", "lo_id", "$1", "$2", "$2"},
		)
		query := `
		INSERT INTO
			%s (%s)
		SELECT
			%s
		FROM
			%s
		WHERE
			topic_id = $3 
		RETURNING lo_id,
			copied_from; `
		stmt := fmt.Sprintf(query, e.TableName(), strings.Join(topicFieldNames, ", "), strings.Join(selectFields, ", "), e.TableName())
		b.Queue(stmt, &newTopicID, &now, &topicID)
	}

	b := &pgx.Batch{}

	for i, topicID := range topicIDs.Elements {
		queue(b, topicID, newTopicIDs.Elements[i])
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	var copiedLo []*entities.CopiedLearningObjective

	for i := 0; i < len(topicIDs.Elements); i++ {
		rows, err := batchResults.Query()
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			e := &entities.CopiedLearningObjective{}
			if err := rows.Scan(&e.LoID, &e.CopiedLoID); err != nil {
				return nil, err
			}
			copiedLo = append(copiedLo, e)
		}
	}
	return copiedLo, nil
}

func (r *LearningObjectiveRepo) UpdateDisplayOrders(ctx context.Context, db database.QueryExecer, mDisplayOrder map[pgtype.Text]pgtype.Int2) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.UpdateDisplayOrders")
	defer span.End()

	queueFn := func(b *pgx.Batch, loID pgtype.Text, displayOrder pgtype.Int2) {
		query := `UPDATE learning_objectives 
		SET display_order = $1, updated_at = NOW()
		WHERE lo_id = $2 AND deleted_at IS NULL`
		b.Queue(query, &displayOrder, &loID)
	}

	var d pgtype.Timestamptz
	d.Set(time.Now())

	b := &pgx.Batch{}
	for loID, displayOrder := range mDisplayOrder {
		queueFn(b, loID, displayOrder)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(mDisplayOrder); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("display_order not changed")
		}
	}

	return nil
}

func (r *LearningObjectiveRepo) RetrieveBookLoByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.BookLearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.RetrieveBookLoByIntervalTime")
	defer span.End()
	stmt := `
	SELECT learning_objectives.%s, book_id FROM learning_objectives
	JOIN topics USING(topic_id)
	JOIN books_chapters USING (chapter_id)
	WHERE topics.deleted_at IS NULL
      AND learning_objectives.deleted_at IS NULL
      AND books_chapters.deleted_at IS NULL
	  AND learning_objectives.deleted_at IS NULL AND learning_objectives.updated_at >= ( now() - $1::interval)`
	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	stmtSelect := fmt.Sprintf(stmt, strings.Join(fields, ", learning_objectives."))
	rows, err := db.Query(ctx, stmtSelect, intervalTime)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	bLearningObjectives := make([]*entities.BookLearningObjective, 0)

	for rows.Next() {
		loTemp := entities.LearningObjective{}
		var (
			bookID pgtype.Text
		)
		scanFields := database.GetScanFields(&loTemp, fields)
		scanFields = append(scanFields, &bookID)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		bLO := entities.BookLearningObjective{
			LearningObjective: loTemp,
			BookID:            bookID,
		}
		bLearningObjectives = append(bLearningObjectives, &bLO)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return bLearningObjectives, nil
}
