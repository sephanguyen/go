package domain

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

// GenFakeMaterials publishedCount must be lower than or equal total
// Generate a slice of LM that contains published and unpublished
func GenFakeLearningMaterials(total int, publishedCount int) (all, published, unpublished []LearningMaterial) {
	materials := make([]LearningMaterial, 0, total)
	publishedMaterials := make([]LearningMaterial, 0, total)
	unpublishedMaterials := make([]LearningMaterial, 0, total)

	for i := 0; i < total; i++ {
		published := false
		if publishedCount > 0 {
			published = true
			publishedCount--
		}
		m := LearningMaterial{
			ID:           idutil.ULIDNow(),
			Name:         fmt.Sprintf("Fake LM %d", i),
			Published:    published,
			DisplayOrder: i,
		}

		materials = append(materials, m)
		if published {
			publishedMaterials = append(publishedMaterials, m)
		} else {
			unpublishedMaterials = append(unpublishedMaterials, m)
		}
	}

	return materials, publishedMaterials, unpublishedMaterials
}
