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
	"go.opencensus.io/trace"
)

type TaskAssignmentRepo struct{}

func (r *TaskAssignmentRepo) Insert(ctx context.Context, db database.QueryExecer, e *entities.TaskAssignment) error {
	ctx, span := interceptors.StartSpan(ctx, "TaskAssignmentRepo.Insert")
	defer span.End()
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func (r *TaskAssignmentRepo) List(ctx context.Context, db database.QueryExecer, learningMaterialIds pgtype.TextArray) ([]*entities.TaskAssignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "TaskAssignmentRepo.List")
	defer span.End()

	b := &entities.TaskAssignment{}
	fieldName, _ := b.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = ANY($1::_TEXT) AND deleted_at IS NULL", strings.Join(fieldName, ", "), b.TableName())
	taskAssignments := entities.TaskAssignments{}
	if err := database.Select(ctx, db, query, learningMaterialIds).ScanAll(&taskAssignments); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return taskAssignments, nil
}

func (m *TaskAssignmentRepo) BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.TaskAssignment) error {
	err := dbeureka.BulkUpsert(ctx, db, bulkInsertQuery, items)
	if err != nil {
		return fmt.Errorf("TaskAssignmentRepo database.BulkInsert error: %s", err.Error())
	}
	return nil
}

func (r *TaskAssignmentRepo) ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.TaskAssignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "TaskAssignmentRepo.ListByTopicIDs")
	defer span.End()
	taskAssignments := &entities.TaskAssignments{}
	e := &entities.TaskAssignment{}
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, db, query, topicIDs).ScanAll(taskAssignments); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return taskAssignments.Get(), nil
}
func (r *TaskAssignmentRepo) Update(ctx context.Context, db database.QueryExecer, m *entities.TaskAssignment) error {
	ctx, span := trace.StartSpan(ctx, "TaskAssignmentRepo.Update")
	defer span.End()

	fields := []string{
		"name",
		"updated_at",
		"attachments",
		"instruction",
		"require_duration",
		"require_complete_date",
		"require_understanding_level",
		"require_correctness",
		"require_attachment",
		"require_assignment_note",
	}

	_, err := database.UpdateFields(ctx, m, db.Exec, "learning_material_id", fields)
	if err != nil {
		return fmt.Errorf("database.UpdateFields: %w", err)
	}
	return nil
}

func (r *TaskAssignmentRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.TaskAssignment) error {
	ctx, span := interceptors.StartSpan(ctx, "TaskAssignmentRepo.Upsert")
	defer span.End()
	fieldNames, values := e.FieldMap()
	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES(%s)
		ON CONFLICT ON CONSTRAINT task_assignment_pk
		DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			display_order = EXCLUDED.display_order,
			updated_at = now(),
			attachments = EXCLUDED.attachments,
			instruction = EXCLUDED.instruction,
			require_duration = EXCLUDED.require_duration,
			require_complete_date = EXCLUDED.require_complete_date,
			require_understanding_level = EXCLUDED.require_understanding_level,
			require_correctness = EXCLUDED.require_correctness,
			require_attachment = EXCLUDED.require_attachment,
			require_assignment_note = EXCLUDED.require_assignment_note
			`,
		e.TableName(),
		strings.Join(fieldNames, ","),
		database.GeneratePlaceholders(len(fieldNames)),
	)
	pgTag, err := db.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	if pgTag.RowsAffected() == 0 {
		return fmt.Errorf("upsert TaskAssignment failed")
	}
	return nil
}
