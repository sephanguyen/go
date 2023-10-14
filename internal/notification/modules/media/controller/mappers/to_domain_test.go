package mappers

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/media/domain"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_PbToMediaDomain(t *testing.T) {
	t.Parallel()
	protoTimeNow := timestamppb.Now()
	comment := &npb.Comment{
		Comment: "comment",
		Duration: &durationpb.Duration{
			Seconds: 10,
		},
	}
	media := &npb.Media{
		MediaId: "media-id",
		Name:    "name",
		Comments: []*npb.Comment{
			comment,
		},
		Resource:  "resource",
		CreatedAt: protoTimeNow,
		UpdatedAt: protoTimeNow,
		Type:      npb.MediaType_MEDIA_TYPE_IMAGE,
		Images:    nil,
		FileSize:  10,
	}

	enComments := []*domain.Comment{
		{
			Comment:  "comment",
			Duration: 10,
		},
	}
	var json pgtype.JSONB
	json.Set(enComments)
	expectMedia := &domain.Media{
		MediaID:         database.Text("media-id"),
		Name:            database.Text("name"),
		Resource:        database.Text("resource"),
		Type:            database.Text(npb.MediaType_MEDIA_TYPE_IMAGE.String()),
		Comments:        json,
		ConvertedImages: pgtype.JSONB{Status: pgtype.Null},
		FileSize:        database.Int8(10),
	}
	expectMedia.DeletedAt.Set(nil)
	expectMedia.CreatedAt.Set(time.Unix(media.CreatedAt.Seconds, int64(media.CreatedAt.Nanos)))
	expectMedia.UpdatedAt.Set(time.Unix(media.UpdatedAt.Seconds, int64(media.UpdatedAt.Nanos)))
	result, err := PbToMediaDomain(media)
	assert.NoError(t, err)
	assert.Equal(t, result, expectMedia)

}
