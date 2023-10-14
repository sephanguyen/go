package learning_objective

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

func (s *Suite) userInsertALearningObjective(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lo := &sspb.LearningObjectiveBase{
		Base: &sspb.LearningMaterialBase{
			TopicId: stepState.TopicIDs[0],
			Name:    "LearningObjective",
		},
	}

	stepState.Response, stepState.ResponseErr = sspb.NewLearningObjectiveClient(s.EurekaConn).InsertLearningObjective(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.InsertLearningObjectiveRequest{
		LearningObjective: lo,
	})
	if stepState.ResponseErr == nil {
		stepState.TopicLODisplayOrderCounter++
		stepState.LearningObjectiveID = stepState.Response.(*sspb.InsertLearningObjectiveResponse).LearningMaterialId
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreLearningObjectivesExistedInTopic(ctx context.Context) (context.Context, error) {
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
		insertLO := &sspb.InsertLearningObjectiveRequest{
			LearningObjective: &sspb.LearningObjectiveBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: stepState.TopicIDs[0],
					Name:    fmt.Sprintf("LO-name+%d", i),
				},
			},
		}
		resp, err := sspb.NewLearningObjectiveClient(s.EurekaConn).InsertLearningObjective(s.AuthHelper.SignedCtx(ctx, stepState.Token), insertLO)
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

func (s *Suite) ourSystemGeneratesACorrectDisplayOrderForLearningObjective(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lo := &entities.LearningObjectiveV2{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.Response.(*sspb.InsertLearningObjectiveResponse).LearningMaterialId),
		},
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(lo), ","), lo.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, lo.ID).ScanOne(lo); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if int32(lo.DisplayOrder.Int) != stepState.TopicLODisplayOrderCounter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect LearningObjective DisplayOrder: expected %d, got %d", stepState.TopicLODisplayOrderCounter, lo.DisplayOrder.Int)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesTopicDisplayOrderCounterOfLearningObjectiveCorrectly(ctx context.Context) (context.Context, error) {
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

func (s *Suite) userInsertALearningObjectiveWithField(ctx context.Context, field string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &sspb.InsertLearningObjectiveRequest{
		LearningObjective: &sspb.LearningObjectiveBase{
			Base: &sspb.LearningMaterialBase{
				TopicId: stepState.TopicIDs[0],
				Name:    "LearningObjective",
			},
		},
	}

	switch field {
	case "manual_grading":
		req.LearningObjective.ManualGrading = true
	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("field %s not belong to LO", field)
	}

	stepState.Response, stepState.ResponseErr = sspb.NewLearningObjectiveClient(s.EurekaConn).InsertLearningObjective(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)

	stepState.Request = req

	if stepState.ResponseErr == nil {
		stepState.TopicLODisplayOrderCounter++
		stepState.LearningObjectiveID = stepState.Response.(*sspb.InsertLearningObjectiveResponse).LearningMaterialId
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustCreateLOWithValue(ctx context.Context, field string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := stepState.Request.(*sspb.InsertLearningObjectiveRequest)

	lo := &entities.LearningObjectiveV2{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.Response.(*sspb.InsertLearningObjectiveResponse).LearningMaterialId),
		},
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(lo), ","), lo.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, lo.ID).ScanOne(lo); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	switch field {
	case "manual_grading":
		if lo.ManualGrading.Bool != req.LearningObjective.ManualGrading {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected LO ManualGrading: want %t, got %t", req.LearningObjective.ManualGrading, lo.ManualGrading.Bool)
		}
	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("field %s not belong to LO", field)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
