package entities

import (
	"github.com/jackc/pgtype"

	"github.com/manabie-com/backend/internal/golibs/database"
)

// BDDInstance maps to e2e_instances in draft database
type BDDInstance struct {
	ID pgtype.Text
	// Duration         pgtype.Int4
	Status           pgtype.Text
	Name             pgtype.Text
	StatusStatistics pgtype.JSONB
	Flavor           pgtype.JSONB
	Tags             pgtype.TextArray
	StartedAt        pgtype.Timestamptz
	EndedAt          pgtype.Timestamptz
}

func (b *BDDInstance) FieldMap() ([]string, []interface{}) {
	return []string{
			"instance_id",
			// "duration",
			"status",
			"name",
			"status_statistics",
			"flavor",
			"tags",
			"started_at",
			"ended_at",
		}, []interface{}{
			&b.ID,
			// &b.Duration,
			&b.Status,
			&b.Name,
			&b.StatusStatistics,
			&b.Flavor,
			&b.Tags,
			&b.StartedAt,
			&b.EndedAt,
		}
}

func (BDDInstance) TableName() string {
	return "e2e_instances"
}

// BDDFeature maps to e2e_features in draft database
type BDDFeature struct {
	ID         pgtype.Text
	InstanceID pgtype.Text
	// Duration   pgtype.Int4
	Status pgtype.Text
	URI    pgtype.Text
	// Data        pgtype.Text
	Keyword pgtype.Text
	Name    pgtype.Text
	// MediaType   pgtype.Text
	// Rules       pgtype.TextArray
	// Description pgtype.Text
	// Scenarios pgtype.JSONB
	// Background  pgtype.JSONB
	// Elements  pgtype.TextArray
	Tags pgtype.TextArray
	// Children  pgtype.JSONB
	StartedAt pgtype.Timestamptz
	EndedAt   pgtype.Timestamptz
}

func (b *BDDFeature) FieldMap() ([]string, []interface{}) {
	return []string{
			"feature_id",
			"instance_id",
			// "duration",
			"status",
			"uri",
			"Keyword",
			"name",
			// "scenarios",
			"tags",
			"started_at",
			"ended_at",
		}, []interface{}{
			&b.ID,
			&b.InstanceID,
			// &b.Duration,
			&b.Status,
			&b.URI,
			&b.Keyword,
			&b.Name,
			// &b.Scenarios,
			&b.Tags,
			&b.StartedAt,
			&b.EndedAt,
		}
}

func (BDDFeature) TableName() string {
	return "e2e_features"
}

// BDDScenario maps to e2e_scenarios in draft database
type BDDScenario struct {
	ID        pgtype.Text
	FeatureID pgtype.Text
	Tags      pgtype.TextArray
	Keyword   pgtype.Text
	Name      pgtype.Text
	Steps     pgtype.JSONB
	Status    pgtype.Text
	Pickle    pgtype.JSONB
	StartedAt pgtype.Timestamptz
	EndedAt   pgtype.Timestamptz
}

func (b *BDDScenario) FieldMap() ([]string, []interface{}) {
	return []string{
			"scenario_id",
			"feature_id",
			"tags",
			"Keyword",
			"name",
			"steps",
			"status",
			"pickle",
			"started_at",
			"ended_at",
		}, []interface{}{
			&b.ID,
			&b.FeatureID,
			&b.Tags,
			&b.Keyword,
			&b.Name,
			&b.Steps,
			&b.Status,
			&b.Pickle,
			&b.StartedAt,
			&b.EndedAt,
		}
}

func (BDDScenario) TableName() string {
	return "e2e_scenarios"
}

// BDDScenario maps to e2e_steps in draft database
type BDDStep struct {
	ID         pgtype.Text
	ScenarioID pgtype.Text
	// Duration   pgtype.Numeric
	Status    pgtype.Text
	URI       pgtype.Text
	Name      pgtype.Text
	Message   pgtype.Text
	StartedAt pgtype.Timestamptz
	EndedAt   pgtype.Timestamptz
}

func (b *BDDStep) FieldMap() ([]string, []interface{}) {
	return []string{
			"step_id",
			"scenario_id",
			// "duration",
			"status",
			"uri",
			"name",
			"message",
			"started_at",
			"ended_at",
		}, []interface{}{
			&b.ID,
			&b.ScenarioID,
			// &b.Duration,
			&b.Status,
			&b.URI,
			&b.Name,
			&b.Message,
			&b.StartedAt,
			&b.EndedAt,
		}
}

func (BDDStep) TableName() string {
	return "e2e_steps"
}

type SkippedBDDTest struct {
	ID           pgtype.Int4
	Repository   pgtype.Varchar
	FeaturePath  pgtype.Varchar
	ScenarioName pgtype.Varchar
	CreatedBy    pgtype.Varchar
	CreatedAt    pgtype.Timestamp
}

func (s *SkippedBDDTest) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"repository",
			"feature_path",
			"scenario_name",
			"created_by",
			"created_at",
		}, []interface{}{
			&s.ID,
			&s.Repository,
			&s.FeaturePath,
			&s.ScenarioName,
			&s.CreatedBy,
			&s.CreatedAt,
		}
}

func (SkippedBDDTest) TableName() string {
	return "skipped_bdd_test"
}

type SkippedBDDTests []*SkippedBDDTest

func (ss *SkippedBDDTests) Add() database.Entity {
	e := &SkippedBDDTest{}
	*ss = append(*ss, e)

	return e
}
