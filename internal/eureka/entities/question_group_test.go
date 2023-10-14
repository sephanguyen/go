package entities

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/s3"
	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestQuestionGroupsToQuestionGroupProtoBufMess(t *testing.T) {
	t.Parallel()
	now := database.Timestamptz(time.Now())
	url, _ := s3.GenerateUploadURL("", "", "rendered rich text")
	tcs := []struct {
		name     string
		input    QuestionGroups
		expected []*cpb.QuestionGroup
	}{
		{
			name: "full field",
			input: QuestionGroups{
				{
					BaseEntity: BaseEntity{
						CreatedAt: now,
						UpdatedAt: now,
						DeletedAt: now,
					},
					QuestionGroupID:    database.Text("group-id-1"),
					LearningMaterialID: database.Text("lm-id-1"),
					Name:               database.Text("name 1"),
					Description:        database.Text("description 1"),
					totalChildren:      database.Int4(2),
					totalPoints:        database.Int4(5),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				},
				{
					BaseEntity: BaseEntity{
						CreatedAt: now,
						UpdatedAt: now,
						DeletedAt: now,
					},
					QuestionGroupID:    database.Text("group-id-2"),
					LearningMaterialID: database.Text("lm-id-1"),
					Name:               database.Text("name 2"),
					Description:        database.Text("description 2"),
					totalChildren:      database.Int4(3),
					totalPoints:        database.Int4(6),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				},
			},
			expected: []*cpb.QuestionGroup{
				{
					QuestionGroupId:    "group-id-1",
					LearningMaterialId: "lm-id-1",
					Name:               "name 1",
					Description:        "description 1",
					CreatedAt:          timestamppb.New(now.Time),
					UpdatedAt:          timestamppb.New(now.Time),
					TotalChildren:      2,
					TotalPoints:        5,
					RichDescription: &cpb.RichText{
						Raw:      "raw rich text",
						Rendered: url,
					},
				},
				{
					QuestionGroupId:    "group-id-2",
					LearningMaterialId: "lm-id-1",
					Name:               "name 2",
					Description:        "description 2",
					CreatedAt:          timestamppb.New(now.Time),
					UpdatedAt:          timestamppb.New(now.Time),
					TotalChildren:      3,
					TotalPoints:        6,
					RichDescription: &cpb.RichText{
						Raw:      "raw rich text",
						Rendered: url,
					},
				},
			},
		},
		{
			name: "miss total children and points field",
			input: QuestionGroups{
				{
					BaseEntity: BaseEntity{
						CreatedAt: now,
						UpdatedAt: now,
						DeletedAt: now,
					},
					QuestionGroupID:    database.Text("group-id-1"),
					LearningMaterialID: database.Text("lm-id-1"),
					Name:               database.Text("name 1"),
					Description:        database.Text("description 1"),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				},
				{
					BaseEntity: BaseEntity{
						CreatedAt: now,
						UpdatedAt: now,
						DeletedAt: now,
					},
					QuestionGroupID:    database.Text("group-id-2"),
					LearningMaterialID: database.Text("lm-id-1"),
					Name:               database.Text("name 2"),
					Description:        database.Text("description 2"),
					RichDescription: database.JSONB(&entities.RichText{
						Raw:         "raw rich text",
						RenderedURL: url,
					}),
				},
			},
			expected: []*cpb.QuestionGroup{
				{
					QuestionGroupId:    "group-id-1",
					LearningMaterialId: "lm-id-1",
					Name:               "name 1",
					Description:        "description 1",
					CreatedAt:          timestamppb.New(now.Time),
					UpdatedAt:          timestamppb.New(now.Time),
					TotalChildren:      0,
					TotalPoints:        0,
					RichDescription: &cpb.RichText{
						Raw:      "raw rich text",
						Rendered: url,
					},
				},
				{
					QuestionGroupId:    "group-id-2",
					LearningMaterialId: "lm-id-1",
					Name:               "name 2",
					Description:        "description 2",
					CreatedAt:          timestamppb.New(now.Time),
					UpdatedAt:          timestamppb.New(now.Time),
					TotalChildren:      0,
					TotalPoints:        0,
					RichDescription: &cpb.RichText{
						Raw:      "raw rich text",
						Rendered: url,
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := QuestionGroupsToQuestionGroupProtoBufMess(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
