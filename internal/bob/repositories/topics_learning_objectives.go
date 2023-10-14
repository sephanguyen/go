package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.opencensus.io/trace"
)

type TopicsLearningObjectivesRepo struct{}

type TopicLearningObjective struct {
	Topic             *entities.Topic
	LearningObjective *entities.LearningObjective
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DisplayOrder      pgtype.Int2
}

func (r *TopicsLearningObjectivesRepo) Create(
	ctx context.Context, db database.QueryExecer,
	m *entities.TopicsLearningObjectives,
) error {
	ctx, span := trace.StartSpan(ctx, "TopicsLearningObjectives.Create")
	defer span.End()

	now := time.Now()
	m.CreatedAt.Set(now)
	m.UpdatedAt.Set(now)
	m.DeletedAt.Set(nil)

	cmd, err := database.Insert(ctx, m, db.Exec)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("cannot insert new " + m.TableName())
	}

	return nil
}

func (r *TopicsLearningObjectivesRepo) BulkImport(
	ctx context.Context, db database.QueryExecer,
	topicsLearningsObjectives []*entities.TopicsLearningObjectives,
) error {
	ctx, span := trace.StartSpan(ctx, "TopicsLearningObjectivesRepo.BulkImport")
	defer span.End()

	queueFn := func(b *pgx.Batch, p *entities.TopicsLearningObjectives) {
		fieldNames := database.GetFieldNames(p)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := `INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT topics_learning_objectives_pk
		DO UPDATE SET updated_at = $5, deleted_at = $6`
		b.Queue(fmt.Sprintf(query, p.TableName(), strings.Join(fieldNames, ","), placeHolders), database.GetScanFields(p, fieldNames)...)
	}

	var d pgtype.Timestamptz
	d.Set(time.Now())

	b := &pgx.Batch{}
	for _, tlp := range topicsLearningsObjectives {
		if tlp.LoID.String == "" {
			return fmt.Errorf("missing lo_id")
		}
		if tlp.TopicID.String == "" {
			return fmt.Errorf("missing topic_id")
		}

		tlp.CreatedAt = d
		tlp.UpdatedAt = d
		tlp.DeletedAt.Set(nil)
		queueFn(b, tlp)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(topicsLearningsObjectives); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("topics learning objectives not inserted")
		}
	}

	return nil
}

func (r *TopicsLearningObjectivesRepo) RetrieveByTopicIDs(
	ctx context.Context, db database.QueryExecer,
	topicIDs pgtype.TextArray,
) ([]*TopicLearningObjective, error) {
	ctx, span := trace.StartSpan(ctx, "TopicsLearningObjectiveRepo.RetrieveByTopicIDs")
	defer span.End()

	topicsLearningObjects := &entities.TopicsLearningObjectives{}
	topic := &entities.Topic{}
	topicFields := database.GetFieldNames(topic)
	learningObjective := &entities.LearningObjective{}
	learningObjectiveFields := database.GetFieldNames(learningObjective)

	stmt := "SELECT t.%s, tlo.created_at, tlo.updated_at, tlo.display_order, lo.%s " +
		"FROM %s AS tlo " +
		"JOIN %s AS t ON tlo.topic_id = t.topic_id " +
		"JOIN %s AS lo ON tlo.lo_id = lo.lo_id " +
		"WHERE tlo.deleted_at IS NULL AND t.deleted_at IS NULL AND lo.deleted_at IS NULL AND tlo.topic_id = ANY($1)"

	query := fmt.Sprintf(stmt, strings.Join(topicFields, ", t."), strings.Join(learningObjectiveFields, ", lo."), topicsLearningObjects.TableName(), topic.TableName(), learningObjective.TableName())

	rows, err := db.Query(ctx, query, topicIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var topicsLearningObjectsResult []*TopicLearningObjective
	for rows.Next() {
		var (
			createdAt    pgtype.Timestamptz
			updatedAt    pgtype.Timestamptz
			displayOrder pgtype.Int2
		)
		t := &entities.Topic{}
		lo := &entities.LearningObjective{}

		scanFields := database.GetScanFields(t, topicFields)
		scanFields = append(scanFields, &createdAt, &updatedAt, &displayOrder)
		scanFields = append(scanFields, database.GetScanFields(lo, learningObjectiveFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		topicsLearningObjects := &TopicLearningObjective{
			Topic:             t,
			LearningObjective: lo,
			DisplayOrder:      displayOrder,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
		}
		topicsLearningObjectsResult = append(topicsLearningObjectsResult, topicsLearningObjects)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return topicsLearningObjectsResult, nil
}

func (r *TopicsLearningObjectivesRepo) RetrieveByLoIDs(
	ctx context.Context, db database.QueryExecer,
	loIDs pgtype.TextArray,
) ([]*TopicLearningObjective, error) {
	ctx, span := trace.StartSpan(ctx, "TopicsLearningObjectiveRepo.RetrieveByLoIDs")
	defer span.End()

	topicsLearningObjects := &entities.TopicsLearningObjectives{}
	topic := &entities.Topic{}
	topicFields := database.GetFieldNames(topic)
	learningObjective := &entities.LearningObjective{}
	learningObjectiveFields := database.GetFieldNames(learningObjective)

	stmt := "SELECT t.%s, tlo.created_at, tlo.updated_at, tlo.display_order, lo.%s " +
		"FROM %s AS tlo " +
		"JOIN %s AS t ON tlo.topic_id = t.topic_id " +
		"JOIN %s AS lo ON tlo.lo_id = lo.lo_id " +
		"WHERE tlo.deleted_at IS NULL AND t.deleted_at IS NULL AND lo.deleted_at IS NULL AND tlo.lo_id = ANY($1)"

	query := fmt.Sprintf(stmt, strings.Join(topicFields, ", t."), strings.Join(learningObjectiveFields, ", lo."), topicsLearningObjects.TableName(), topic.TableName(), learningObjective.TableName())

	rows, err := db.Query(ctx, query, loIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var topicsLearningObjectsResult []*TopicLearningObjective
	for rows.Next() {
		var (
			createdAt    pgtype.Timestamptz
			updatedAt    pgtype.Timestamptz
			displayOrder pgtype.Int2
		)
		t := &entities.Topic{}
		lo := &entities.LearningObjective{}

		scanFields := database.GetScanFields(t, topicFields)
		scanFields = append(scanFields, &createdAt, &updatedAt, &displayOrder)
		scanFields = append(scanFields, database.GetScanFields(lo, learningObjectiveFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		topicsLearningObjects := &TopicLearningObjective{
			Topic:             t,
			LearningObjective: lo,
			DisplayOrder:      displayOrder,
			CreatedAt:         createdAt,
			UpdatedAt:         updatedAt,
		}
		topicsLearningObjectsResult = append(topicsLearningObjectsResult, topicsLearningObjects)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return topicsLearningObjectsResult, nil
}

func (r *TopicsLearningObjectivesRepo) SoftDeleteByTopicIDAndLoID(
	ctx context.Context, db database.QueryExecer,
	topicID, loID pgtype.Text) error {
	topicLearningObjective := &entities.TopicsLearningObjectives{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE topic_id = $1 AND lo_id = $2`, topicLearningObjective.TableName())
	cmd, err := db.Exec(ctx, sql, topicID, loID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows were affected")
	}

	return nil
}

func (r *TopicsLearningObjectivesRepo) SoftDeleteByTopicID(
	ctx context.Context, db database.QueryExecer,
	topicID pgtype.Text) error {
	topicLearningObjective := &entities.TopicsLearningObjectives{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE topic_id = $1`, topicLearningObjective.TableName())
	cmd, err := db.Exec(ctx, sql, topicID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows were affected")
	}

	return nil
}

func (r *TopicsLearningObjectivesRepo) SoftDeleteByLoIDs(
	ctx context.Context, db database.QueryExecer,
	loIDs pgtype.TextArray) error {
	topicLearningObjective := &entities.TopicsLearningObjectives{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE lo_id = ANY($1::TEXT[]) AND deleted_at IS NULL`, topicLearningObjective.TableName())
	if _, err := db.Exec(ctx, sql, loIDs); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (r *TopicsLearningObjectivesRepo) BulkUpdateDisplayOrder(
	ctx context.Context, db database.QueryExecer,
	topicsLearningsObjectives []*entities.TopicsLearningObjectives,
) error {
	ctx, span := trace.StartSpan(ctx, "TopicsLearningObjectivesRepo.BulkUpdateDisplayOrder")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *entities.TopicsLearningObjectives) {
		query := fmt.Sprintf("UPDATE %s SET display_order = $1, updated_at = now() WHERE lo_id = $2 AND topic_id = $3 AND deleted_at IS NULL", e.TableName())
		b.Queue(query, e.DisplayOrder, e.LoID, e.TopicID)
	}
	b := &pgx.Batch{}
	for _, each := range topicsLearningsObjectives {
		queueFn(b, each)
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
