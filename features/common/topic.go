package common

import (
	"context"
	"fmt"

	bob_constant "github.com/manabie-com/backend/internal/bob/constants"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) AListOfValidTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	s.aListOfValidChaptersInDB(ctx)

	s.BobDB.Exec(ctx, `
	(chapter_id, name, country, subject, grade, display_order, updated_at, created_at)
	VALUES
	('chapter_id_1', 'chapter_id_1', 'COUNTRY_VN', 'SUBJECT_BIOLOGY', 12, 1, now(), now()),
	('chapter_id_2', 'chapter_id_2', 'COUNTRY_VN', 'SUBJECT_BIOLOGY', 12, 1, now(), now()),
	('chapter_id_3', 'chapter_id_3', 'COUNTRY_VN', 'SUBJECT_BIOLOGY', 12, 1, now(), now())
	ON CONFLICT DO NOTHING;`)
	if stepState.ChapterID == "" {
		if ctx, err := s.insertAChapter(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve ")
		}
	}
	t1 := s.generateValidTopic(stepState.ChapterID)
	t1.ChapterId = "chapter_id_1"
	t1.Type = epb.TopicType_TOPIC_TYPE_LEARNING
	stepState.Topics = append(stepState.Topics, &t1)

	t2 := s.generateValidTopic(stepState.ChapterID)
	t2.ChapterId = "chapter_id_2"
	t2.Type = epb.TopicType_TOPIC_TYPE_EXAM
	stepState.Topics = append(stepState.Topics, &t2)

	t3 := s.generateValidTopic(stepState.ChapterID)
	t3.ChapterId = "chapter_id_3"
	t3.Type = epb.TopicType_TOPIC_TYPE_PRACTICAL
	stepState.Topics = append(stepState.Topics, &t3)

	stepState.Request = &epb.UpsertTopicsRequest{Topics: []*epb.Topic{&t1, &t2, &t3}}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateValidTopic(chapterID string) epb.Topic {
	return epb.Topic{
		Id:           s.newID(),
		Name:         "topic 1",
		Country:      epb.Country_COUNTRY_VN,
		Grade:        "G12",
		Subject:      epb.Subject_SUBJECT_MATHS,
		Type:         epb.TopicType_TOPIC_TYPE_LEARNING,
		CreatedAt:    timestamppb.Now(),
		UpdatedAt:    timestamppb.Now(),
		Status:       epb.TopicStatus_TOPIC_STATUS_NONE,
		DisplayOrder: 1,
		PublishedAt:  timestamppb.Now(),
		SchoolId:     bob_constant.ManabieSchool,
		IconUrl:      "topic-icon",
		ChapterId:    chapterID,
	}
}

func (s *suite) AdminInsertsAListOfValidTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.EurekaConn).Upsert(
		s.SignedCtx(ctx), stepState.Request.(*epb.UpsertTopicsRequest))
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}
