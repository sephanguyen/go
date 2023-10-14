package flashcard

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

func (s *Suite) userInsertsAFlashcard(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	insertFlashcardtReq := &sspb.InsertFlashcardRequest{
		Flashcard: &sspb.FlashcardBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    "assignment-name",
			},
		},
	}
	resp, err := sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard(s.AuthHelper.SignedCtx((ctx), stepState.Token), insertFlashcardtReq)

	stepState.Request = insertFlashcardtReq
	stepState.Response = resp
	stepState.ResponseErr = err
	stepState.TopicLODisplayOrderCounter++
	stepState.FlashcardID = resp.GetLearningMaterialId()
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userInsertsALmsv2Flashcard(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	insertFlashcardtReq := &sspb.InsertFlashcardRequest{
		Flashcard: &sspb.FlashcardBase{
			Base: &sspb.LearningMaterialBase{
				TopicId:    stepState.TopicIDs[0],
				Name:       "flashcard lmsv2",
				VendorType: sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY,
			},
		},
	}
	resp, err := sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard(s.AuthHelper.SignedCtx((ctx), stepState.Token), insertFlashcardtReq)

	stepState.Request = insertFlashcardtReq
	stepState.Response = resp
	stepState.ResponseErr = err
	stepState.TopicLODisplayOrderCounter++
	stepState.FlashcardID = resp.GetLearningMaterialId()
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreFlashcardsExistedInTopic(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := rand.Intn(5) + 3 // maximum = 8
	wg := &sync.WaitGroup{}
	var err error
	cErrs := make(chan error, n)
	cIDs := make(chan string, n)

	defer func() {
		close(cErrs)
		close(cIDs)
	}()
	genAndInsert := func(ctx context.Context, i int, wg *sync.WaitGroup) {
		defer wg.Done()
		insertFlashcardtReq := &sspb.InsertFlashcardRequest{
			Flashcard: &sspb.FlashcardBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: stepState.TopicIDs[0],
					Name:    fmt.Sprintf("flashcard-name+%d", i),
				},
			},
		}
		resp, err := sspb.NewFlashcardClient(s.EurekaConn).InsertFlashcard(s.AuthHelper.SignedCtx(ctx, stepState.Token), insertFlashcardtReq)
		if err != nil {
			cErrs <- err
		}
		cIDs <- resp.LearningMaterialId
	}
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go genAndInsert(ctx, i, wg)
	}
	go func() {
		wg.Wait()
	}()
	go func() {

	}()
	for i := 0; i < n; i++ {
		select {
		case errTemp := <-cErrs:
			err = multierr.Combine(err, errTemp)
		case id := <-cIDs:
			stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, id)
			stepState.TopicLODisplayOrderCounter++
		}
	}
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemGeneratesACorrectDisplayOrderForFlashcard(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	fc := &entities.Flashcard{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.Response.(*sspb.InsertFlashcardResponse).LearningMaterialId),
		},
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(fc), ","), fc.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, fc.ID).ScanOne(fc); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if int32(fc.DisplayOrder.Int) != stepState.TopicLODisplayOrderCounter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect Flashcard DisplayOrder: expected %d, got %d", stepState.TopicLODisplayOrderCounter, fc.DisplayOrder.Int)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesTopicLODisplayOrderCounterCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	topic := &entities.Topic{
		ID: database.Text(stepState.TopicIDs[0]),
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE topic_id = $1", strings.Join(database.GetFieldNames(topic), ","), topic.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, topic.ID).ScanOne(topic); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if topic.LODisplayOrderCounter.Int != stepState.TopicLODisplayOrderCounter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect Topic LODisplayOrderCounter: expected %d, got %d", stepState.TopicLODisplayOrderCounter, topic.LODisplayOrderCounter.Int)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theLmsv2FlashcardIsCreatedWithCorrectData(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	flashcardID := stepState.FlashcardID
	res := &entities.LearningMaterial{}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(res), ","), res.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, flashcardID).ScanOne(res); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if res.ID.String != flashcardID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect learning material id: expected %s, got %s", res.ID.String, flashcardID)
	}
	if res.Type.String != sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String() {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect learning material type: expected %s, got %s", res.Type.String, sspb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD.String())
	}
	if res.VendorType.String != sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String() {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect learning material vendor type: expected %s, got %s", res.VendorType.String, sspb.LearningMaterialVendorType_LM_VENDOR_TYPE_LEARNOSITY.String())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
