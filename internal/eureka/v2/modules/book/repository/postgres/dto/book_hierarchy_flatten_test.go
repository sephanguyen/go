package dto

import (
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
)

func TestBookHierarchyFlatten_ToEntity(t *testing.T) {
	// arrange
	dto := BookHierarchyFlatten{
		BookID:             database.Text("b_id"),
		ChapterID:          database.Text("chapter_id"),
		TopicID:            database.Text("topic_id"),
		LearningMaterialID: database.Text("learning_material_id"),
	}

	// act
	res := dto.ToEntity()

	// assert
	assert.Equal(t, domain.BookHierarchyFlatten{
		BookID:             "b_id",
		ChapterID:          "chapter_id",
		TopicID:            "topic_id",
		LearningMaterialID: "learning_material_id",
	}, res)
}
