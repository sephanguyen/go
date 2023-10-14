package entities

import (
	"github.com/manabie-com/backend/internal/bob/entities"
)

type PresetStudyPlan struct {
	TableName struct{} `sql:"preset_study_plans"`
	entities.PresetStudyPlan
}
