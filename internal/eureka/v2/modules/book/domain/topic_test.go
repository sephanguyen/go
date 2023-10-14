package domain

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/stretchr/testify/assert"
)

func TestTopic_RemoveUnpublishedMaterials(t *testing.T) {
	t.Run("Remove all when all learning materials are unpublished", func(t *testing.T) {
		t.Parallel()
		// arrange
		_, _, unpublished := GenFakeLearningMaterials(10, 5)
		topic := Topic{
			ID:                "ID 1",
			Name:              "Name 1",
			LearningMaterials: unpublished,
		}
		expectedTopic := Topic{
			ID:                topic.ID,
			Name:              topic.Name,
			LearningMaterials: []LearningMaterial{},
		}
		// act
		topic.RemoveUnpublishedMaterials()

		// assert
		assert.Equal(t, expectedTopic, topic)

	})

	t.Run("Remove a set of unpublished", func(t *testing.T) {
		t.Parallel()
		// arrange
		all, published, _ := GenFakeLearningMaterials(10, 5)
		topic := Topic{
			ID:                "ID 1",
			Name:              "Name 1",
			LearningMaterials: all,
		}
		expectedTopic := Topic{
			ID:                topic.ID,
			Name:              topic.Name,
			LearningMaterials: published,
		}
		// act
		topic.RemoveUnpublishedMaterials()

		// assert
		assert.Equal(t, expectedTopic, topic)
	})

	t.Run("Remove nothing while all learning materials are published", func(t *testing.T) {
		t.Parallel()
		// arrange
		_, published, _ := GenFakeLearningMaterials(10, 5)
		topic := Topic{
			ID:                "ID 1",
			Name:              "Name 1",
			LearningMaterials: published,
		}
		expectedTopic := Topic{
			ID:                topic.ID,
			Name:              topic.Name,
			LearningMaterials: published,
		}
		// act
		topic.RemoveUnpublishedMaterials()

		// assert
		assert.Equal(t, expectedTopic, topic)
	})
}

// GenFakeTopics publishedCount must be lower than or equal total
func GenFakeTopics(total int, publishedCount int) (all, published, unpublished []Topic) {
	all = make([]Topic, 0, total)
	published = make([]Topic, 0, total)
	unpublished = make([]Topic, 0, total)

	for i := 0; i < total; i++ {
		var topic Topic
		if publishedCount > 0 {
			publishedCount--
			_, publishedMaterials, _ := GenFakeLearningMaterials(3, 3)
			topic = Topic{
				ID:                idutil.ULIDNow(),
				Name:              fmt.Sprintf("Fake Topic %d", i),
				DisplayOrder:      i,
				LearningMaterials: publishedMaterials,
			}
			published = append(published, topic)
		} else {
			_, _, unpublishedMaterials := GenFakeLearningMaterials(3, 0)
			topic = Topic{
				ID:                idutil.ULIDNow(),
				Name:              fmt.Sprintf("Fake Topic %d", i),
				DisplayOrder:      i,
				LearningMaterials: unpublishedMaterials,
			}
			unpublished = append(unpublished, topic)
		}

		all = append(all, topic)
	}

	return all, published, unpublished
}
