package mappers

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/modules/media/domain"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"go.uber.org/multierr"
)

func PbToCommentDomain(comments []*npb.Comment) []*domain.Comment {
	result := make([]*domain.Comment, 0, len(comments))
	for _, comment := range comments {
		c := &domain.Comment{
			Comment:  comment.Comment,
			Duration: comment.Duration.GetSeconds(),
		}
		result = append(result, c)
	}
	return result
}

func PbToMediaDomain(src *npb.Media) (*domain.Media, error) {
	dst := &domain.Media{}
	database.AllNullEntity(dst)
	if src.MediaId == "" {
		src.MediaId = idutil.ULIDNow()
	}
	comments := PbToCommentDomain(src.Comments)

	err := multierr.Combine(
		dst.MediaID.Set(src.MediaId),
		dst.Resource.Set(src.Resource),
		dst.Name.Set(src.Name),
		dst.Comments.Set(comments),
		dst.Type.Set(src.Type.String()),
		dst.FileSize.Set(src.FileSize),
	)
	if src.CreatedAt != nil {
		_ = multierr.Append(err, dst.CreatedAt.Set(time.Unix(src.CreatedAt.Seconds, int64(src.CreatedAt.Nanos))))
	} else {
		_ = multierr.Append(err, dst.CreatedAt.Set(time.Now()))
	}

	if src.UpdatedAt != nil {
		_ = multierr.Append(err, dst.UpdatedAt.Set(time.Unix(src.UpdatedAt.Seconds, int64(src.UpdatedAt.Nanos))))
	} else {
		_ = multierr.Append(err, dst.UpdatedAt.Set(time.Now()))
	}
	return dst, err
}
