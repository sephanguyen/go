package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aListOfValidTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	bookResp, err := epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &epb.UpsertBooksRequest{
		Books: s.generateBooks(1, nil),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	chapterResp, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(contextWithToken(s, ctx), &epb.UpsertChaptersRequest{
		Chapters: s.generateChapters(ctx, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	stepState.ChapterID = chapterResp.ChapterIds[0]
	topics := s.generateTopics(ctx, 3, nil)
	topics[0].Type = epb.TopicType_TOPIC_TYPE_LEARNING
	topics[1].Type = epb.TopicType_TOPIC_TYPE_EXAM
	topics[2].Type = epb.TopicType_TOPIC_TYPE_PRACTICAL
	stepState.Topics = topics
	stepState.Request = &epb.UpsertTopicsRequest{Topics: topics}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminInsertsAListOfValidTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(
		s.signedCtx(ctx), stepState.Request.(*epb.UpsertTopicsRequest))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp

	stepState.TopicIDs = resp.GetTopicIds()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateValidTopic(chapterID string) epb.Topic {
	return epb.Topic{
		Id:           idutil.ULIDNow(),
		Name:         "topic 1",
		Country:      epb.Country_COUNTRY_VN,
		Grade:        "G12",
		Subject:      epb.Subject_SUBJECT_BIOLOGY,
		Type:         epb.TopicType_TOPIC_TYPE_ASSIGNMENT,
		CreatedAt:    timestamppb.Now(),
		UpdatedAt:    timestamppb.Now(),
		Status:       epb.TopicStatus_TOPIC_STATUS_DRAFT,
		DisplayOrder: 1,
		PublishedAt:  timestamppb.Now(),
		SchoolId:     constants.ManabieSchool,
		IconUrl:      "topic-icon",
		ChapterId:    chapterID,
	}
}
