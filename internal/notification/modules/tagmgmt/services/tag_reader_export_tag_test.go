package services

import (
	"context"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/tagmgmt/repositories"
	"github.com/manabie-com/backend/mock/testutil"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ExportTags(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	tagRepo := &mock_repositories.MockTagRepo{}
	svc := &TagMgmtReaderService{
		DB:      mockDB.DB,
		TagRepo: tagRepo,
	}
	tags := entities.Tags{
		&entities.Tag{
			TagID:      database.Text("tag-id-1"),
			TagName:    database.Text("tag-name-1"),
			IsArchived: database.Bool(false),
		},
	}
	sb := strings.Builder{}
	sb.WriteString("\"tag_id\",\"tag_name\",\"is_archived\"\n\"tag-id-1\",\"tag-name-1\",\"0\"\n")
	base64Data := []byte(sb.String())
	t.Run("happy case", func(t *testing.T) {
		ctx := context.Background()
		tagRepo.On("FindByFilter", ctx, mockDB.DB, mock.Anything).Once().Return(tags, uint32(0), nil)

		res, err := svc.ExportTags(
			ctx,
			&npb.ExportTagsRequest{},
		)
		assert.Nil(t, err)
		assert.Equal(t, base64Data, res.Data)
	})
}
