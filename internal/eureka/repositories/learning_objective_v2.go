package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LearningObjectiveRepoV2 struct{}

func (r *LearningObjectiveRepoV2) Insert(ctx context.Context, db database.QueryExecer, e *entities.LearningObjectiveV2) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.Insert")
	defer span.End()
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func (r *LearningObjectiveRepoV2) Update(ctx context.Context, db database.QueryExecer, e *entities.LearningObjectiveV2) error {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.Update")
	defer span.End()
	if _, err := database.UpdateFields(ctx, e, db.Exec, "learning_material_id", []string{
		"name",
		"updated_at",
		"video",
		"study_guide",
		"video_script",
	}); err != nil {
		return fmt.Errorf("database.UpdateFields: %w", err)
	}
	return nil
}

func (r *LearningObjectiveRepoV2) ListByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjectiveV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.ListByIDs")
	defer span.End()
	los := &entities.LearningObjectiveV2s{}
	e := &entities.LearningObjectiveV2{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = ANY($1::_TEXT) AND deleted_at IS NULL", strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, db, query, ids).ScanAll(los); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return los.Get(), nil
}

func (r *LearningObjectiveRepoV2) ListLOBaseByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjectiveBaseV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepo.ListLOBaseByIDs")
	defer span.End()
	bases := make([]*entities.LearningObjectiveBaseV2, 0)
	e := &entities.LearningObjectiveV2{}

	stmt := fmt.Sprintf(`
		SELECT lo.%s, array_length(qs.quiz_external_ids, 1) as total_question
		FROM %s lo
		LEFT JOIN quiz_sets qs ON qs.lo_id = lo.learning_material_id AND qs.deleted_at IS NULL
		WHERE lo.learning_material_id = ANY($1::_TEXT) AND lo.deleted_at IS NULL
	`,
		strings.Join(database.GetFieldNames(e), ", lo."),
		e.TableName(),
	)

	rows, err := db.Query(ctx, stmt, ids)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		b := &entities.LearningObjectiveBaseV2{}

		_, values := b.LearningObjectiveV2.FieldMap()
		values = append(values, &b.TotalQuestion)

		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		bases = append(bases, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return bases, nil
}

func (r *LearningObjectiveRepoV2) ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjectiveV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "LearningObjectiveRepoV2.ListByTopicIDs")
	defer span.End()
	los := &entities.LearningObjectiveV2s{}
	e := &entities.LearningObjectiveV2{}
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, db, query, topicIDs).ScanAll(los); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return los.Get(), nil
}

func (m *LearningObjectiveRepoV2) BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.LearningObjectiveV2) error {
	err := dbeureka.BulkUpsert(ctx, db, bulkInsertQuery, items)
	if err != nil {
		return fmt.Errorf("LearningObjectiveRepoV2 database.BulkInsert error: %s", err.Error())
	}
	return nil
}

func (r *LearningObjectiveRepoV2) RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjectiveV2, error) {
	learningObjective := &entities.LearningObjectiveV2{}
	fieldNames := database.GetFieldNames(learningObjective)
	query := `SELECT lo.%s FROM %s lo
		INNER JOIN learning_material lm ON lm.learning_material_id = lo.learning_material_id
		WHERE ($1::TEXT[] IS NULL OR lo.topic_id = ANY($1::TEXT[]))
		AND lo.deleted_at IS NULL
		AND lm.is_published = TRUE
		ORDER BY lo.display_order ASC
	`

	learningObjectives := entities.LearningObjectiveV2s{}
	err := database.Select(ctx, db, fmt.Sprintf(query, strings.Join(fieldNames, ", lo."), learningObjective.TableName()), &topicIDs).ScanAll(&learningObjectives)
	if err != nil {
		return nil, err
	}

	return learningObjectives, nil
}

func (r *LearningObjectiveRepoV2) CountLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) (int, error) {
	learningObjective := &entities.LearningObjectiveV2{}
	var count int
	query := `SELECT COUNT(*) FROM %s lo
		INNER JOIN learning_material lm ON lm.learning_material_id = lo.learning_material_id 
		WHERE ($1::TEXT[] IS NULL OR lo.topic_id = ANY($1::TEXT[]))
		AND lo.deleted_at IS NULL
		AND lm.is_published = TRUE`

	err := db.QueryRow(ctx, fmt.Sprintf(query, learningObjective.TableName()), &topicIDs).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
