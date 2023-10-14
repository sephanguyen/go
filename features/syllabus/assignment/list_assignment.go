package assignment

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"go.uber.org/multierr"
)

func (s *Suite) thereAreAssignmentsExisted(ctx context.Context) (context.Context, error) {
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
		insertAssignmentReq := &sspb.InsertAssignmentRequest{
			Assignment: &sspb.AssignmentBase{
				Base: &sspb.LearningMaterialBase{
					TopicId: stepState.TopicIDs[0],
					Name:    fmt.Sprintf("assignment-name+%d", i),
				},
			},
		}
		resp, err := sspb.NewAssignmentClient(s.EurekaConn).InsertAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), insertAssignmentReq)
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

func (s *Suite) userListAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ListAssignmentReq := &sspb.ListAssignmentRequest{
		LearningMaterialIds: stepState.LearningMaterialIDs,
	}
	stepState.Response, stepState.ResponseErr = sspb.NewAssignmentClient(s.EurekaConn).ListAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Token), ListAssignmentReq)
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnAssignmentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	response := stepState.Response.(*sspb.ListAssignmentResponse)
	if len(response.Assignments) != len(stepState.LearningMaterialIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("incorrect number of assignments, expected %d, got %d", len(stepState.LearningMaterialIDs), len(response.Assignments))
	}

	for _, assignment := range response.Assignments {
		if !golibs.InArrayString(assignment.Base.LearningMaterialId, stepState.LearningMaterialIDs) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected assignment id: %q in list %v of assignments ids: %q", assignment.Base.LearningMaterialId, stepState.LearningMaterialIDs, response.Assignments)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
