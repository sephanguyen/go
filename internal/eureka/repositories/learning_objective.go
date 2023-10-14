package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LearningObjectiveRepo struct{}

func (r *LearningObjectiveRepo) RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.RetrieveByTopicIDs")
	defer span.End()

	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	query := fmt.Sprintf("SELECT DISTINCT lo.%s FROM %s lo LEFT JOIN %s t ON t.topic_id = lo.topic_id WHERE lo.topic_id = ANY($1) AND lo.deleted_at IS NULL AND t.deleted_at IS NULL", strings.Join(fields, ", lo."), lo.TableName(), (&entities.Topic{}).TableName())
	pp := entities.LearningObjectives{}
	if err := database.Select(ctx, db, query, &topicIds).ScanAll(&pp); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return pp, nil
}

func (r *LearningObjectiveRepo) RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.RetrieveByTopicIDs")
	defer span.End()

	lo := &entities.LearningObjective{}
	fields := database.GetFieldNames(lo)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lo_id = ANY($1::_TEXT) AND deleted_at IS NULL", strings.Join(fields, ","), lo.TableName())
	pp := entities.LearningObjectives{}
	if err := database.Select(ctx, db, query, &ids).ScanAll(&pp); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
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

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("row.Err: %w", err)
		}
	}
	return copiedLo, nil
}

func (r *LearningObjectiveRepo) BulkImport(ctx context.Context, db database.QueryExecer, learningObjectives []*entities.LearningObjective) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.BulkImport")
	defer span.End()

	queueFn := func(b *pgx.Batch, p *entities.LearningObjective) {
		fieldNames := database.GetFieldNames(p)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := `INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT learning_objectives_pk
		DO UPDATE SET 
			name = $2,
			country = $3,
			grade = $4,
			subject = $5,
			topic_id = $6,
			master_lo_id = $7,
			video_script = $9,
			prerequisites = $10,
			video = $11,
			study_guide = $12,
			school_id = $13,
			updated_at = $15,
			type = $17,
			instruction = $18,
			grade_to_pass = $19,
			manual_grading = $20,
			time_limit = $21,
			maximum_attempt = $22,
			approve_grading = $23,
			grade_capping = $24,
			review_option = $25,
			vendor_type = $26
		`
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
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("learning objectives not inserted")
		}
	}

	return nil
}

func (r *LearningObjectiveRepo) SoftDeleteWithLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.SoftDeleteWithLoIDs")
	defer span.End()
	query := `UPDATE learning_objectives SET deleted_at = NOW() WHERE lo_id = ANY($1::TEXT[]) AND deleted_at IS NULL`
	tag, err := db.Exec(ctx, query, &loIDs)
	if err != nil {
		return 0, fmt.Errorf("err db.Exec: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *LearningObjectiveRepo) RetrieveBookLoByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.BookLearningObjective, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.RetrieveBookLoByIntervalTime")
	defer span.End()
	stmt := `
	SELECT learning_objectives.%s, book_id, books_chapters.chapter_id, topics.topic_id FROM learning_objectives
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
			bookID, chapterID, topicID pgtype.Text
		)
		scanFields := database.GetScanFields(&loTemp, fields)
		scanFields = append(scanFields, &bookID, &chapterID, &topicID)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		bLO := entities.BookLearningObjective{
			LearningObjective: loTemp,
			BookID:            bookID,
			ChapterID:         chapterID,
			TopicID:           topicID,
		}
		bLearningObjectives = append(bLearningObjectives, &bLO)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row.Err: %w", err)
	}

	return bLearningObjectives, nil
}

func (r *LearningObjectiveRepo) RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjective, error) {
	learningObjective := &entities.LearningObjective{}
	fieldNames := database.GetFieldNames(learningObjective)
	query := `
		SELECT
			%s
		FROM
			%s
		WHERE
			topic_id = ANY($1)
			AND deleted_at IS NULL
		`
	var learningObjectives entities.LearningObjectives
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fieldNames, ","), learningObjective.TableName()), &topicIDs).ScanAll(&learningObjectives)
	if err != nil {
		return nil, err
	}
	return learningObjectives, nil
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

func (r *LearningObjectiveRepo) CountTotal(ctx context.Context, db database.QueryExecer) (*pgtype.Int8, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.CountTotal")
	defer span.End()

	lo := &entities.LearningObjective{}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", lo.TableName())
	var result pgtype.Int8
	if err := db.QueryRow(ctx, query).Scan(&result); err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return &result, nil
}

func (r *LearningObjectiveRepo) UpdateName(ctx context.Context, db database.QueryExecer, loID pgtype.Text, name pgtype.Text) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.UpdateName")
	defer span.End()
	e := &entities.LearningObjective{}

	query := fmt.Sprintf("UPDATE %s SET name = $1, updated_at = now() WHERE lo_id = $2::TEXT AND deleted_at IS NULL", e.TableName())
	cmd, err := db.Exec(ctx, query, name, loID)

	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}

	return cmd.RowsAffected(), nil
}
