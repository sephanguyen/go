package dto

import (
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/stretchr/testify/assert"
)

func TestLearningMaterialDto_ToEntity(t *testing.T) {
	// arrange
	dto := LearningMaterialDto{
		ID:           database.Text("id"),
		TopicID:      database.Text("topic_id"),
		Name:         database.Text("name"),
		Type:         database.Text("LEARNING_MATERIAL_LEARNING_OBJECTIVE"),
		DisplayOrder: database.Int2(1),
		IsPublished:  database.Bool(true),
	}

	// act
	res := dto.ToEntity()

	// assert
	assert.Equal(t, domain.LearningMaterial{
		ID:           "id",
		TopicID:      "topic_id",
		Name:         "name",
		Type:         constants.LearningObjective,
		DisplayOrder: 1,
		Published:    true,
	}, res)
}
