package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.opencensus.io/trace"
	"go.uber.org/multierr"
)

type TopicsAssignmentsRepo struct {
}

func (r *TopicsAssignmentsRepo) Upsert(
	ctx context.Context, db database.QueryExecer,
	m *entities.TopicsAssignments,
) error {
	ctx, span := trace.StartSpan(ctx, "TopicsAssignments.Upsert")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		m.CreatedAt.Set(now),
		m.UpdatedAt.Set(now),
		m.DeletedAt.Set(nil),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	fieldNames := database.GetFieldNames(m)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := `INSERT INTO %s (%s) VALUES (%s)
	ON CONFLICT ON CONSTRAINT topics_assignments_pk 
	DO UPDATE SET created_at = NOW(), updated_at = NOW(), deleted_at = NULL`
	query = fmt.Sprintf(query, m.TableName(), strings.Join(fieldNames, ","), placeHolders)
	args := database.GetScanFields(m, fieldNames)

	commandTag, err := db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows were affected")
	}
	return nil
}

func (r *TopicsAssignmentsRepo) BulkUpsert(
	ctx context.Context, db database.QueryExecer,
	topicsAssignmentsList []*entities.TopicsAssignments,
) error {
	ctx, span := trace.StartSpan(ctx, "TopicsAssignments.BulkUpsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, p *entities.TopicsAssignments) {
		fieldNames := database.GetFieldNames(p)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))
		query := `INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT topics_assignments_pk
		DO UPDATE SET updated_at = $5, deleted_at = $6`
		b.Queue(fmt.Sprintf(query, p.TableName(), strings.Join(fieldNames, ","), placeHolders), database.GetScanFields(p, fieldNames)...)
	}

	var d pgtype.Timestamptz
	_ = d.Set(time.Now())
	b := &pgx.Batch{}
	for _, ta := range topicsAssignmentsList {
		if ta.TopicID.String == "" {
			return fmt.Errorf("missing topic_id")
		}
		if ta.AssignmentID.String == "" {
			return fmt.Errorf("missing assignment_id")
		}
		ta.CreatedAt = d
		ta.UpdatedAt = d
		_ = ta.DeletedAt.Set(nil)
		queueFn(b, ta)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < len(topicsAssignmentsList); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("topics assignments not inserted")
		}
	}
	return nil
}

func (r *TopicsAssignmentsRepo) RetrieveByAssignmentIDs(
	ctx context.Context, db database.QueryExecer, assignmentIDs []string,
) ([]*entities.TopicsAssignments, error) {
	ctx, span := trace.StartSpan(ctx, "TopicsAssignmentsRepo.RetrieveByAssignmentIDs")
	defer span.End()

	topicAssignment := &entities.TopicsAssignments{}
	topicAssignmentFields := database.GetFieldNames(topicAssignment)

	stmt := `SELECT t.%s 
		FROM %s AS t 
		WHERE t.deleted_at IS NULL AND t.assignment_id = ANY($1)`

	query := fmt.Sprintf(stmt, strings.Join(topicAssignmentFields, ", t."), topicAssignment.TableName())

	rows, err := db.Query(ctx, query, &assignmentIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var topicsAssignments []*entities.TopicsAssignments
	for rows.Next() {
		topicAssignment := new(entities.TopicsAssignments)

		if err := rows.Scan(database.GetScanFields(topicAssignment, topicAssignmentFields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		topicsAssignments = append(topicsAssignments, topicAssignment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return topicsAssignments, nil
}

func (r *TopicsAssignmentsRepo) SoftDeleteByAssignmentIDs(
	ctx context.Context, db database.QueryExecer,
	assignmentIDs pgtype.TextArray) error {
	topicsAssignments := &entities.TopicsAssignments{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = NOW() WHERE assignment_id = ANY($1) AND deleted_at IS NULL`, topicsAssignments.TableName())
	cmd, err := db.Exec(ctx, sql, assignmentIDs)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows were affected")
	}

	return nil
}

func (r *TopicsAssignmentsRepo) BulkUpdateDisplayOrder(
	ctx context.Context, db database.QueryExecer,
	topicsAssignments []*entities.TopicsAssignments,
) error {
	ctx, span := trace.StartSpan(ctx, "TopicsLearningObjectivesRepo.BulkUpdateDisplayOrder")
	defer span.End()

	queueFn := func(b *pgx.Batch, e *entities.TopicsAssignments) {
		query := fmt.Sprintf("UPDATE %s SET display_order = $1, updated_at = now() WHERE assignment_id = $2 AND topic_id = $3 AND deleted_at IS NULL", e.TableName())
		b.Queue(query, e.DisplayOrder, e.AssignmentID, e.TopicID)
	}
	b := &pgx.Batch{}
	for _, each := range topicsAssignments {
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
