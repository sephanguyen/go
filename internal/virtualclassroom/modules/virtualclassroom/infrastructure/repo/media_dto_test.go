package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMediaFromEntity(t *testing.T) {
	t.Parallel()
	now := time.Time{}
	tcs := []struct {
		name  string
		media *domain.Media
		dto   *Media
	}{
		{
			name: "full fields",
			media: &domain.Media{
				ID:        "media-id-1",
				Name:      "name 1",
				Resource:  "resource-1",
				CreatedAt: now,
				UpdatedAt: now,
				Type:      domain.MediaTypeImage,
			},
			dto: &Media{
				MediaID:   database.Text("media-id-1"),
				Name:      database.Text("name 1"),
				Resource:  database.Text("resource-1"),
				Type:      database.Text(string(domain.MediaTypeImage)),
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewMediaFromEntity(tc.media)
			require.NoError(t, err)
			assert.Equal(t, tc.dto.MediaID, actual.MediaID)
			assert.Equal(t, tc.dto.Name, actual.Name)
			assert.Equal(t, tc.dto.Resource, actual.Resource)
			assert.Equal(t, tc.dto.Type, actual.Type)
			assert.Equal(t, tc.dto.CreatedAt, actual.CreatedAt)
			assert.Equal(t, tc.dto.UpdatedAt, actual.UpdatedAt)
		})
	}
}
