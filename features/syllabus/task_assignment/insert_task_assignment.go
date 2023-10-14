package task_assignment

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

func (s *Suite) userInsertAValidTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := &sspb.InsertTaskAssignmentRequest{
		TaskAssignment: &sspb.TaskAssignmentBase{
			Base: &sspb.LearningMaterialBase{
				Name:    "Task Assignment",
				TopicId: stepState.TopicID,
			},
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = sspb.NewTaskAssignmentClient(s.EurekaConn).InsertTaskAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	if stepState.ResponseErr == nil {
		stepState.TopicLODisplayOrderCounter++
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreTaskAssignmentsExistedInTopic(ctx context.Context) (context.Context, error) {
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
		insertTaskAssignmentReq := &sspb.InsertTaskAssignmentRequest{
			TaskAssignment: &sspb.TaskAssignmentBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: stepState.TopicID,
					Name:    fmt.Sprintf("task-assignment-lo-name+%d", i),
				},
			},
		}
		resp, err := sspb.NewTaskAssignmentClient(s.EurekaConn).InsertTaskAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), insertTaskAssignmentReq)
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

func (s *Suite) taskAssignmentMustBeCreated(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.InsertTaskAssignmentResponse)
	e := &entities.TaskAssignment{}

	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE learning_material_id = $1", e.TableName())
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, &resp.LearningMaterialId).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if count != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of task assignment %d, got %d", 1, count)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemGeneratesACorrectDisplayOrderForTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	e := &entities.TaskAssignment{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.Response.(*sspb.InsertTaskAssignmentResponse).LearningMaterialId),
		},
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(e), ","), e.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, e.ID).ScanOne(e); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if int32(e.DisplayOrder.Int) != stepState.TopicLODisplayOrderCounter {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect TaskAssignment DisplayOrder: expected %d, got %d", stepState.TopicLODisplayOrderCounter, e.DisplayOrder.Int)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesTopicLODisplayOrderCounterCorrectlyWithNewTaskAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	topic := &entities.Topic{
		ID: database.Text(stepState.TopicID),
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
