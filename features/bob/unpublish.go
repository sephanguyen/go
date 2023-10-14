package bob

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/bob/entities"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *suite) theLearnerUnpublish(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.AuthToken, err = s.generateExchangeToken(stepState.studentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).Unpublish(s.signedCtx(ctx), &bpb.UnpublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.studentID})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) noRecordIndicatingThatTheLearnerIsUnpublishAnUploadStreamInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, isPublish, err := s.checkLearnerIsPublish(ctx, stepState.studentID)
	if err != nil {
		return ctx, err

	}
	if isPublish {
		return ctx, fmt.Errorf("none unpublishing streaming")
	}
	return ctx, nil
}
func (s *suite) theNumberOfStreamOfTheLessonHaveToEqualToNumber(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if numberOfStream != stepState.numberOfStream {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the number of stream have to equal default")
	}

	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) unpublishReturnsTheResponse(ctx context.Context, args string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	resp, ok := (stepState.Response).(*bpb.UnpublishResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Response must be *pb.UnpublishResponse")
	}
	if resp.Status.String() != args {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected status of response: expected :%s, got: %s", args, resp.Status)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) twoLearnerPreparedPublishInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.firstLearner = stepState.StudentIds[0]
	stepState.secondLearner = stepState.StudentIds[1]
	stepState.studentID = stepState.firstLearner
	for i := 0; i < 2; i++ {
		ctx, err := s.theLearnerPrepareToPublish(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("theLearnerPrepareToPublish: %w", err)
		}
		stepState.studentID = stepState.secondLearner
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) recordIndicatingTheTwoLearnerPreparedPublishInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, isPublish1, err := s.checkLearnerIsPublish(ctx, stepState.firstLearner)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	ctx, isPublish2, err := s.checkLearnerIsPublish(ctx, stepState.secondLearner)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if !isPublish1 || !isPublish2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("both learner have to publish streaming")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) twoLearnersUnpublishInConcurrently(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		var err error
		firstToken, err := s.generateExchangeToken(stepState.firstLearner, entities.UserGroupStudent)
		if err != nil {
			stepState.firstResponseErr = err
			return
		}

		stepState.firstResponse, stepState.firstResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).Unpublish(
			s.signedCtxWithToken(ctx, firstToken),
			&bpb.UnpublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.firstLearner})
	}()
	go func() {
		defer wg.Done()
		var err error
		secondToken, err := s.generateExchangeToken(stepState.secondLearner, entities.UserGroupStudent)
		if err != nil {
			stepState.secondResponseErr = err
			return
		}

		stepState.secondResponse, stepState.secondResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).Unpublish(
			s.signedCtxWithToken(ctx, secondToken),
			&bpb.UnpublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.secondLearner})
	}()
	wg.Wait()
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) noRecordIndicatingThatTheLearnerPreparedToPublishAnUploadStreamInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, isPublish1, err := s.checkLearnerIsPublish(ctx, stepState.firstLearner)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	ctx, isPublish2, err := s.checkLearnerIsPublish(ctx, stepState.secondLearner)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if isPublish1 || isPublish2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("have to no record indicating two learner prepared to publish")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsOKForTheOne(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.firstResponseErr != nil {
		return ctx, stepState.firstResponseErr

	}
	if stepState.secondResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.secondResponseErr
	}
	resp1, ok := (stepState.firstResponse).(*bpb.UnpublishResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Response must be *pb.UnpublishResponse")
	}
	resp2, ok := (stepState.secondResponse).(*bpb.UnpublishResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Response must be *pb.UnpublishResponse")
	}
	if resp1 == resp2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("two responses of requests have to not equal")
	}
	if resp1.Status.String() == resp2.Status.String() && resp2.Status.String() != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("both request not returns OK")
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLearnerUnpublishTwiceInConcurrent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		var err error
		token, err := s.generateExchangeToken(stepState.studentID, entities.UserGroupStudent)
		if err != nil {
			stepState.firstResponseErr = err
			return
		}

		stepState.firstResponse, stepState.firstResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).Unpublish(
			s.signedCtxWithToken(ctx, token),
			&bpb.UnpublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.studentID})
	}()
	go func() {
		defer wg.Done()
		var err error
		token, err := s.generateExchangeToken(stepState.studentID, entities.UserGroupStudent)
		if err != nil {
			stepState.secondResponseErr = err
			return
		}

		stepState.secondResponse, stepState.secondResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).Unpublish(
			s.signedCtxWithToken(ctx, token),
			&bpb.UnpublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.studentID})
	}()
	wg.Wait()
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) unpublishReturnsTheResponseForAnother(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.firstResponseErr != nil {
		return ctx, stepState.firstResponseErr

	}
	if stepState.secondResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.secondResponseErr
	}
	resp1, ok := (stepState.firstResponse).(*bpb.UnpublishResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Response must be *pb.UnpublishResponse")
	}
	resp2, ok := (stepState.secondResponse).(*bpb.UnpublishResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Response must be *pb.UnpublishResponse")
	}
	if resp1.Status.String() != arg1 && resp2.Status.String() != arg1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting req1 %s || req2 %s, got %s status code", resp1.Status.String(), resp2.Status.String(), arg1)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theNumberOfStreamOfTheLessonHaveToDecrease(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if numberOfStream != stepState.numberOfStream {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the number of stream have to decrease 1, expected %d, got %d", stepState.numberOfStream, numberOfStream)
	}
	return StepStateToContext(ctx, stepState), nil
}
