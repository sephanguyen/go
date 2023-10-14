package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"go.opencensus.io/trace"
	"go.uber.org/multierr"
)

type GeneralAssignmentRepo struct {
}

func (r *GeneralAssignmentRepo) Insert(
	ctx context.Context, db database.QueryExecer,
	m *entities.GeneralAssignment,
) error {
	ctx, span := trace.StartSpan(ctx, "GeneralAssignmentRepo.Insert")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		m.CreatedAt.Set(now),
		m.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}
	fieldNames := database.GetFieldNames(m)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := `INSERT INTO %s (%s) VALUES (%s);`
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

func (r *GeneralAssignmentRepo) Update(
	ctx context.Context, db database.QueryExecer,
	m *entities.GeneralAssignment,
) error {
	ctx, span := trace.StartSpan(ctx, "GeneralAssignmentRepo.Update")
	defer span.End()

	fields := []string{
		"name",
		"attachments",
		"max_grade",
		"instruction",
		"is_required_grade",
		"allow_resubmission",
		"require_attachment",
		"allow_late_submission",
		"require_assignment_note",
		"require_video_submission",
		"updated_at",
	}

	cmdTag, err := database.UpdateFields(ctx, m, db.Exec, "learning_material_id", fields)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update general assignment")
	}
	return nil
}

const queryListByTopicIDs = "SELECT %s FROM %s WHERE topic_id = ANY($1::_TEXT) AND deleted_at IS NULL"

func (r *GeneralAssignmentRepo) ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.GeneralAssignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "GeneralAssignmentRepo.ListByTopicIDs")
	defer span.End()
	generalAssignments := &entities.GeneralAssignments{}
	gAssignment := &entities.GeneralAssignment{}
	query := fmt.Sprintf(queryListByTopicIDs, strings.Join(database.GetFieldNames(gAssignment), ","), gAssignment.TableName())
	if err := database.Select(ctx, db, query, topicIDs).ScanAll(generalAssignments); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}
	return generalAssignments.Get(), nil
}

const bulkInsertQuery = `INSERT INTO %s (%s) VALUES %s`

func (m *GeneralAssignmentRepo) BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.GeneralAssignment) error {
	err := dbeureka.BulkUpsert(ctx, db, bulkInsertQuery, items)
	if err != nil {
		return fmt.Errorf("GeneralAssignmentRepo database.BulkInsert error: %s", err.Error())
	}
	return nil
}

func (r *GeneralAssignmentRepo) List(ctx context.Context, db database.QueryExecer, learningMaterialIds pgtype.TextArray) ([]*entities.GeneralAssignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.List")
	defer span.End()

	b := &entities.GeneralAssignment{}
	fieldName, _ := b.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = ANY($1::_TEXT) AND deleted_at IS NULL", strings.Join(fieldName, ", "), b.TableName())
	assignments := entities.GeneralAssignments{}
	if err := database.Select(ctx, db, query, learningMaterialIds).ScanAll(&assignments); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return assignments, nil
}
