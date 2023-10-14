package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *suite) schoolAdminHasCreatedSomeTopicsBefore(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	numberOfTopic := rand.Intn(5) + 1
	reqTopics := s.generateTopics(ctx, numberOfTopic, nil)
	res, err := s.upsertTopics(ctx, reqTopics)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.OldTopicIDs = res.TopicIds
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminHasCreateSomeTopics(ctx context.Context, typeTopic string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	numberOfTopics := rand.Intn(5) + 2
	reqTopics := s.generateTopics(ctx, numberOfTopics, nil)

	if typeTopic == "new and old" {
		stepState.NumberOfUpdatedOldTopics = rand.Intn(numberOfTopics)
		// Assert at least 1 old topic
		if stepState.NumberOfUpdatedOldTopics == 0 {
			stepState.NumberOfUpdatedOldTopics = 1
		}
		// Assert at least 1 new topic
		if stepState.NumberOfUpdatedOldTopics == numberOfTopics {
			stepState.NumberOfUpdatedOldTopics = numberOfTopics - 1
		}
		// Assert no overcounted old topics
		if stepState.NumberOfUpdatedOldTopics > len(stepState.OldTopicIDs) {
			stepState.NumberOfUpdatedOldTopics = len(stepState.OldTopicIDs)
		}
		for i := 0; i < stepState.NumberOfUpdatedOldTopics; i++ {
			reqTopics[i].Id = stepState.OldTopicIDs[i]
		}
	}

	res, err := s.upsertTopics(ctx, reqTopics)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TopicIDs = res.TopicIds
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToSaveTheTopicsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	chapterRepo := &repositories.ChapterRepo{}
	chapter, err := chapterRepo.FindByID(ctx, s.DB, database.Text(stepState.ChapterID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to get the chapter: %w", err)
	}
	var expectedCurrentTopicDisplayOrder = len(stepState.OldTopicIDs) + len(stepState.TopicIDs) - stepState.NumberOfUpdatedOldTopics

	if int(chapter.CurrentTopicDisplayOrder.Int) != expectedCurrentTopicDisplayOrder {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong current topic display order on chapter: want: %d, actual: %d", expectedCurrentTopicDisplayOrder, int(chapter.CurrentTopicDisplayOrder.Int))
	}

	topicRepo := &repositories.TopicRepo{}
	topicIDs := make([]string, 0)
	topicIDs = append(topicIDs, stepState.OldTopicIDs...)
	topicIDs = append(topicIDs, stepState.TopicIDs...)

	topics, err := topicRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(topicIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to find topics: %w", err)
	}
	mapTopics := convertTopics2MapTopics(topics)
	for _, c := range topics {
		if int32(c.DisplayOrder.Int) > chapter.CurrentTopicDisplayOrder.Int || int32(c.DisplayOrder.Int) <= 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong display order on topic")
		}
	}
	// check duplicate
	if isDuplicateTopics(mapTopics) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Duplicate display order on upsert topic")
	}
	return StepStateToContext(ctx, stepState), nil
}

// case: concurrency
func (s *suite) schoolAdminHasCreatedSomeTopics(ctx context.Context, schoolAdminOrdinal string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	numOfTopic := rand.Intn(5) + 2
	pbTopics := s.generateTopics(ctx, numOfTopic, nil)

	if schoolAdminOrdinal == "first" {
		stepState.AuthToken = stepState.AnotherSchoolAdminToken
	} else if schoolAdminOrdinal == "second" {
		stepState.AuthToken = stepState.SchoolAdminToken
	}
	ctx = contextWithToken(s, ctx)

	res, err := s.upsertTopics(ctx, pbTopics)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if schoolAdminOrdinal == "first" {
		stepState.TopicIDs = res.TopicIds
	} else if schoolAdminOrdinal == "second" {
		stepState.AnotherTopicIDs = res.TopicIds
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) twoSchoolAdminCreateSomeTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		Errs error
	)
	reqError := make(chan error, 3)
	wgDone := make(chan bool)
	wg := sync.WaitGroup{}

	wg.Add(2) // 1 for wgDone

	go func() {
		defer wg.Done()
		var err error
		ctx, err = s.schoolAdminHasCreatedSomeTopics(ctx, "first")
		if err != nil {
			reqError <- fmt.Errorf("First school admin upsert topic fail: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		var err error
		ctx, err = s.schoolAdminHasCreatedSomeTopics(ctx, "second")
		if err != nil {
			reqError <- fmt.Errorf("Second school admin upsert topic fail: %w", err)
		}
	}()

	wg.Wait()
	close(wgDone)

ReadError:
	for {
		select {
		case <-wgDone:
			break ReadError
		case err := <-reqError:
			Errs = multierr.Append(Errs, err)
		}
	}

	if Errs != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Failed to upsert topic: %w", Errs)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToStoreTheTopicsInConcurrencyCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	topicRepo := &repositories.TopicRepo{}
	topicIDs := make([]string, 0)
	topicIDs = append(topicIDs, stepState.TopicIDs...)
	topicIDs = append(topicIDs, stepState.AnotherTopicIDs...)

	topics, err := topicRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(topicIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to find topics: %w", err)
	}
	mapTopics := convertTopics2MapTopics(topics)

	chapterRepo := &repositories.ChapterRepo{}
	chapter, err := chapterRepo.FindByID(ctx, s.DB, database.Text(stepState.ChapterID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to get the chapter: %w", err)
	}
	var expectedCurrentTopic = len(topicIDs)
	if int(chapter.CurrentTopicDisplayOrder.Int) != expectedCurrentTopic {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong current topic display order on chapter: want: %d, actual: %d", expectedCurrentTopic, int(chapter.CurrentTopicDisplayOrder.Int))
	}

	topicIDsFirst := stepState.TopicIDs
	topicIDsSecond := stepState.AnotherTopicIDs
	topicsFirst, err := topicRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(topicIDsFirst))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to find first topics: %w", err)
	}
	mapTopicsFirst := convertTopics2MapTopics(topicsFirst)
	topicsSecond, err := topicRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(topicIDsSecond))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to find second topics: %w", err)
	}
	mapTopicsSecond := convertTopics2MapTopics(topicsSecond)

	// check which TopicIds upsert first
	if len(topicsFirst) < 1 || len(topicsSecond) < 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong test when upsert chapter")
	}
	var isTopicIDsFirstUpsertFirst bool
	if mapTopicsFirst[topicIDsFirst[0]].DisplayOrder.Int < mapTopicsSecond[topicIDsSecond[0]].DisplayOrder.Int {
		isTopicIDsFirstUpsertFirst = true
	}
	if isTopicIDsFirstUpsertFirst {
		for _, c := range mapTopicsSecond {
			if c.DisplayOrder.Int < mapTopicsFirst[topicIDsFirst[0]].DisplayOrder.Int {
				return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong display order when upsert topic concurrency")
			}
		}
	} else {
		for _, c := range mapTopicsFirst {
			if c.DisplayOrder.Int < mapTopicsSecond[topicIDsSecond[0]].DisplayOrder.Int {
				return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong display order when upsert topic concurrency")
			}
		}
	}
	// check duplicate
	if isDuplicateTopics(mapTopics) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Duplicate display order on upsert topic")
	}
	return StepStateToContext(ctx, stepState), nil
}

func convertTopics2MapTopics(topics []*entities.Topic) map[string]*entities.Topic {
	mapTopics := make(map[string]*entities.Topic)
	for _, topic := range topics {
		mapTopics[topic.ID.String] = topic
	}
	return mapTopics
}

func isDuplicateTopics(mapTopicss map[string]*entities.Topic) bool {
	mapUniqueDisplayOrder := make(map[int]bool)
	for _, c := range mapTopicss {
		mapUniqueDisplayOrder[int(c.DisplayOrder.Int)] = true
	}
	return len(mapUniqueDisplayOrder) < len(mapTopicss)
}

func (s *suite) schoolAdminHasCreatedSomeTopicsBeforeByOldFlow(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	numberOfTopic := rand.Intn(5) + 1
	reqTopics := s.generateTopics(ctx, numberOfTopic, &epb.Topic{
		Name:         "Topic",
		Subject:      epb.Subject_SUBJECT_BIOLOGY,
		Type:         epb.TopicType_TOPIC_TYPE_LEARNING,
		DisplayOrder: 1,
		TotalLos:     1,
		ChapterId:    stepState.ChapterID,
		SchoolId:     stepState.SchoolIDInt,
	})
	res, err := s.upsertTopics(ctx, reqTopics)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.OldTopicIDs = res.TopicIds

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToSaveTopicsOnBothOldAndNewFlowCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	chapterRepo := &repositories.ChapterRepo{}
	chapter, err := chapterRepo.FindByID(ctx, s.DB, database.Text(stepState.ChapterID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to get the chapter: %w", err)
	}
	var expectedCurrentTopicDisplayOrder = len(stepState.OldTopicIDs) + len(stepState.TopicIDs) - stepState.NumberOfUpdatedOldTopics

	if int(chapter.CurrentTopicDisplayOrder.Int) != expectedCurrentTopicDisplayOrder {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong current topic display order on chapter: want: %d, actual: %d", expectedCurrentTopicDisplayOrder, int(chapter.CurrentTopicDisplayOrder.Int))
	}

	topicRepo := &repositories.TopicRepo{}
	topicIDs := make([]string, 0)
	topicIDs = append(topicIDs, stepState.OldTopicIDs...)
	topicIDs = append(topicIDs, stepState.TopicIDs...)

	topics, err := topicRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(topicIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Unable to find topics: %w", err)
	}
	mapTopics := convertTopics2MapTopics(topics)
	for _, c := range topics {
		if int32(c.DisplayOrder.Int) > chapter.CurrentTopicDisplayOrder.Int || int32(c.DisplayOrder.Int) <= 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong display order on topic")
		}
	}
	if int(mapTopics[stepState.TopicIDs[0]].DisplayOrder.Int) < len(stepState.OldTopicIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Display order on new flow have to larger than number of topics old flow")
	}
	for _, tID := range stepState.OldTopicIDs {
		if mapTopics[tID].DisplayOrder.Int > int16(chapter.CurrentTopicDisplayOrder.Int) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Wrong display order on topic in new flow")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) schoolAdminHasCreateSomeNewTopics(ctx context.Context) (context.Context, error) {
	return s.schoolAdminHasCreateSomeTopics(ctx, "new")
}
