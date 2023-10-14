package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgtype"

	"github.com/manabie-com/backend/internal/draft/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type BDDSuite struct{}

func (b *BDDSuite) AddInstance(ctx context.Context, db database.QueryExecer, e *entities.BDDInstance) error {
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func (b *BDDSuite) MarkInstanceEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDInstance) error {
	if _, err := database.UpdateFields(
		ctx,
		e,
		db.Exec,
		"instance_id",
		[]string{"status", "status_statistics", "ended_at"},
	); err != nil {
		return fmt.Errorf("database.UpdateFields: %v", err)
	}
	return nil
}

func (b *BDDSuite) AddFeature(ctx context.Context, db database.QueryExecer, e *entities.BDDFeature) (pgtype.Text, error) {
	cmdTag, err := database.InsertIgnoreConflict(ctx, e, db.Exec)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("database.Insert: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		var featureID pgtype.Text
		row := db.QueryRow(ctx, "SELECT feature_id FROM e2e_features WHERE instance_id = $1 AND uri = $2", &e.InstanceID, &e.URI)
		if err := row.Scan(&featureID); err != nil {
			return pgtype.Text{}, fmt.Errorf("row.Scan: %v", err)
		}

		return featureID, nil
	}
	return e.ID, nil
}

func (b *BDDSuite) MarkFeatureEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDFeature) error {
	if _, err := database.UpdateFields(
		ctx,
		e,
		db.Exec,
		"feature_id",
		[]string{"status", "ended_at"},
	); err != nil {
		return fmt.Errorf("database.UpdateFields: %v", err)
	}
	return nil
}

func (b *BDDSuite) SetFeatureStatus(ctx context.Context, db database.QueryExecer, e *entities.BDDFeature) error {
	if _, err := database.UpdateFields(
		ctx,
		e,
		db.Exec,
		"feature_id",
		[]string{"status"},
	); err != nil {
		return fmt.Errorf("database.UpdateFields: %v", err)
	}
	return nil
}

func (b *BDDSuite) AddScenario(ctx context.Context, db database.QueryExecer, e *entities.BDDScenario) error {
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %w", err)
	}
	return nil
}

func (b *BDDSuite) MarkScenarioEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDScenario) error {
	if _, err := database.UpdateFields(
		ctx,
		e,
		db.Exec,
		"scenario_id",
		[]string{"status", "ended_at"},
	); err != nil {
		return fmt.Errorf("database.UpdateFields: %v", err)
	}
	return nil
}

func (b *BDDSuite) AddStep(ctx context.Context, db database.QueryExecer, e *entities.BDDStep) error {
	if _, err := database.Insert(ctx, e, db.Exec); err != nil {
		return fmt.Errorf("database.Insert: %v", err)
	}
	return nil
}

func (b *BDDSuite) MarkStepEnded(ctx context.Context, db database.QueryExecer, e *entities.BDDStep) error {
	if _, err := database.UpdateFields(
		ctx,
		e,
		db.Exec,
		"step_id",
		[]string{"status", "message", "ended_at"},
	); err != nil {
		return fmt.Errorf("database.UpdateFields: %v", err)
	}
	return nil
}

func (b *BDDSuite) RetrieveSkippedBDDTestsByRepository(ctx context.Context, db database.QueryExecer, repo pgtype.Varchar) ([]*entities.SkippedBDDTest, error) {
	e := &entities.SkippedBDDTest{}
	fieldNames := database.GetFieldNames(e)
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE repository = $1`, strings.Join(fieldNames, ","), e.TableName())

	var ee entities.SkippedBDDTests
	if err := database.Select(ctx, db, query, repo).ScanAll(&ee); err != nil {
		return nil, fmt.Errorf("database.Select: %v", err)
	}
	return ee, nil
}
