package mappers

import (
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func TagEntitiesToTagsByFilterResponse(tags entities.Tags) []*npb.GetTagsByFilterResponse_Tag {
	if len(tags) == 0 {
		return nil
	}

	tagsByFilter := []*npb.GetTagsByFilterResponse_Tag{}
	for _, tag := range tags {
		tagResponse := npb.GetTagsByFilterResponse_Tag{
			TagId: tag.TagID.String,
			Name:  tag.TagName.String,
		}
		tagsByFilter = append(tagsByFilter, &tagResponse)
	}
	return tagsByFilter
}
