package mappers

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/tagmgmt/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
)

func Test_TagEntitiesToTagsByFilterResponse(t *testing.T) {
	t.Parallel()
	tag1 := &entities.Tag{}
	tag2 := &entities.Tag{}
	database.AllRandomEntity(tag1)
	database.AllRandomEntity(tag2)
	testCases := []struct {
		Name    string
		Tags    entities.Tags
		ExpcRes []*npb.GetTagsByFilterResponse_Tag
	}{
		{
			Name: "happy case",
			Tags: func() entities.Tags {
				tags := entities.Tags{}
				tags = append(tags, tag1)
				tags = append(tags, tag2)
				return tags
			}(),
			ExpcRes: func() []*npb.GetTagsByFilterResponse_Tag {
				res := []*npb.GetTagsByFilterResponse_Tag{}
				res = append(res, &npb.GetTagsByFilterResponse_Tag{
					TagId: tag1.TagID.String,
					Name:  tag1.TagName.String,
				})
				res = append(res, &npb.GetTagsByFilterResponse_Tag{
					TagId: tag2.TagID.String,
					Name:  tag2.TagName.String,
				})
				return res
			}(),
		},
		{
			Name: "nil",
			Tags: func() entities.Tags {
				tags := entities.Tags{}
				return tags
			}(),
			ExpcRes: nil,
		},
	}

	for _, tc := range testCases {
		res := TagEntitiesToTagsByFilterResponse(tc.Tags)
		assert.Equal(t, tc.ExpcRes, res)
	}
}
