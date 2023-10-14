package eureka

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/golibs/constants"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/proto"
)

func (s *suite) generateTopic(ctx context.Context, template *epb.Topic) *epb.Topic {
	stepState := StepStateFromContext(ctx)
	if template == nil {
		template = &epb.Topic{
			SchoolId:  constants.ManabieSchool,
			Subject:   epb.Subject_SUBJECT_BIOLOGY,
			Name:      "Topic",
			ChapterId: stepState.ChapterID,
			Status:    epb.TopicStatus_TOPIC_STATUS_NONE,
			Type:      epb.TopicType_TOPIC_TYPE_LEARNING,
			TotalLos:  1,
		}
	}

	return proto.Clone(template).(*epb.Topic)
}

func (s *suite) generateTopics(ctx context.Context, numberOfBooks int, template *epb.Topic) []*epb.Topic {
	books := make([]*epb.Topic, 0)
	for i := 0; i < numberOfBooks; i++ {
		books = append(books, s.generateTopic(ctx, template))
	}
	return books
}

func (s *suite) upsertTopics(ctx context.Context, topics []*epb.Topic) (*epb.UpsertTopicsResponse, error) {
	res, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(ctx, &epb.UpsertTopicsRequest{
		Topics: topics,
	})
	if err != nil {
		return nil, fmt.Errorf("Can not create topics: %w", err)
	}
	if len(res.TopicIds) != len(topics) {
		return nil, fmt.Errorf("Unexpected number of topics: want %d, got %d", len(topics), len(res.TopicIds))
	}
	return res, nil
}

func (s *suite) userHasCreatedSomeValidTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfTopics := rand.Intn(5) + 2
	reqTopics := s.generateTopics(ctx, numberOfTopics, nil)

	res, err := s.upsertTopics(ctx, reqTopics)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TopicIDs = res.TopicIds
	return StepStateToContext(ctx, stepState), nil
}
