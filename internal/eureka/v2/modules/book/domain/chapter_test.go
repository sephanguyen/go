package domain

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/stretchr/testify/assert"
)

func TestChapter_RemoveUnpublishedTopics(t *testing.T) {
	t.Run("Remove all unpublished", func(t *testing.T) {
		t.Parallel()
		// arrange
		_, _, topics := GenFakeTopics(10, 0)
		chapter := Chapter{
			ID:     "Chapter ID",
			Name:   "Chapter Name",
			Topics: topics,
		}
		expectedChapter := Chapter{
			ID:     chapter.ID,
			Name:   chapter.Name,
			Topics: []Topic{},
		}
		// act
		chapter.RemoveUnpublishedTopics()

		// assert
		assert.Equal(t, expectedChapter, chapter)
	})

	t.Run("Remove a set of unpublished", func(t *testing.T) {
		t.Parallel()
		// arrange
		all, published, _ := GenFakeTopics(10, 5)
		chapter := Chapter{
			ID:     "Chapter ID",
			Name:   "Chapter Name",
			Topics: all,
		}
		expectedChapter := Chapter{
			ID:     chapter.ID,
			Name:   chapter.Name,
			Topics: published,
		}
		// act
		chapter.RemoveUnpublishedTopics()

		// assert
		assert.Equal(t, expectedChapter, chapter)
	})

	t.Run("Remove nothing while all topics are published", func(t *testing.T) {
		t.Parallel()
		// arrange
		_, published, _ := GenFakeTopics(10, 5)
		chapter := Chapter{
			ID:     "Chapter ID",
			Name:   "Chapter Name",
			Topics: published,
		}
		expectedChapter := Chapter{
			ID:     chapter.ID,
			Name:   chapter.Name,
			Topics: published,
		}
		// act
		chapter.RemoveUnpublishedTopics()

		// assert
		assert.Equal(t, expectedChapter, chapter)
	})
}

// GenFakeChapters publishedCount must be lower than or equal total
func GenFakeChapters(total int, publishedCount int) (all, published, unpublished []Chapter) {
	all = make([]Chapter, 0, total)
	published = make([]Chapter, 0, total)
	unpublished = make([]Chapter, 0, total)

	for i := 0; i < total; i++ {
		var chapter Chapter
		if publishedCount > 0 {
			publishedCount--
			_, pubTopics, _ := GenFakeTopics(3, 3)
			chapter = Chapter{
				ID:           idutil.ULIDNow(),
				Name:         fmt.Sprintf("Fake Chapter %d", i),
				DisplayOrder: i,
				Topics:       pubTopics,
			}
			published = append(published, chapter)
		} else {
			_, _, unPubTopics := GenFakeTopics(3, 0)
			chapter = Chapter{
				ID:           idutil.ULIDNow(),
				Name:         fmt.Sprintf("Fake Chapter %d", i),
				DisplayOrder: i,
				Topics:       unPubTopics,
			}
			unpublished = append(unpublished, chapter)
		}

		all = append(all, chapter)
	}

	return all, published, unpublished
}
