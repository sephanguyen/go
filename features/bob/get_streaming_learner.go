package bob

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"github.com/manabie-com/backend/internal/bob/entities"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
)

func (s *suite) learnerPreparedPublishInTheLesson(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rawNumberOfStudent, err := strconv.ParseInt(arg1, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	stepState.numberOfLearners = int(rawNumberOfStudent)
	for i := 0; i < stepState.numberOfLearners; i++ {
		stepState.studentID = stepState.StudentIds[i]
		ctx, err := s.theLearnerPrepareToPublish(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("theLearnerPrepareToPublish: %w", err)
		}
		stepState.studentID = stepState.secondLearner
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) getStreamingLearners(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccountV2(ctx, "student")
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonReaderServiceClient(s.Conn).GetStreamingLearners(s.signedCtx(ctx), &bpb.GetStreamingLearnersRequest{LessonId: stepState.lessonID})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnLearnerIdsWhoAreCurrentlyUploadingInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	resp, ok := (stepState.Response).(*bpb.GetStreamingLearnersResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Response must be *pb.GetStreamingLearnersResponse")
	}
	if len(resp.LearnerIds) != stepState.numberOfLearners {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected length of learner uploading, expected %d, got %d", stepState.numberOfLearners, len(resp.LearnerIds))
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aLessonWithArbitraryNumberOfStudentPublishing(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.numberOfStudentsArePublishing = rand.Intn(s.Cfg.Agora.MaximumLearnerStreamings-5-2) + 5                                       // 5 is min number student are publishing (sub 2 for minimum 2 slots for 2 variables value below)
	stepState.numberOfStudentsWantToPublish = rand.Intn(s.Cfg.Agora.MaximumLearnerStreamings-stepState.numberOfStudentsArePublishing-1) + 1 // sub 1 for below variable
	stepState.numberOfStudentsWantToUnpublish = rand.Intn(s.Cfg.Agora.MaximumLearnerStreamings - stepState.numberOfStudentsArePublishing - stepState.numberOfStudentsWantToPublish + 1)
	// publishing

	for i := 0; i < stepState.numberOfStudentsArePublishing; i++ {
		token, _ := s.generateExchangeToken(stepState.StudentIds[i], entities.UserGroupStudent)
		_, err := bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(s.signedCtxWithToken(ctx, token), &bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.StudentIds[i]})
		if err != nil {
			return StepStateToContext(ctx, stepState), err

		}
	}
	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) studentsPublishAndUnpublishAsTheSameTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		idxPublish   = stepState.numberOfStudentsArePublishing + 1
		idxUnPublish int
		Errs         error
	)
	reqError := make(chan error, 3)
	wgDone := make(chan bool)
	wg := sync.WaitGroup{}

	wg.Add(stepState.numberOfStudentsWantToPublish + stepState.numberOfStudentsWantToUnpublish) //1 for wgDone

	for i := 0; i < stepState.numberOfStudentsWantToPublish || i < stepState.numberOfStudentsWantToUnpublish; i++ {
		if i < stepState.numberOfStudentsWantToPublish {
			go func(i int) {
				defer wg.Done()
				token, _ := s.generateExchangeToken(stepState.StudentIds[i], entities.UserGroupStudent)
				_, err := bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(s.signedCtxWithToken(ctx, token), &bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.StudentIds[i]})
				if err != nil {
					reqError <- err
				}
			}(i + idxPublish)
		}
		if i < stepState.numberOfStudentsWantToUnpublish {
			go func(i int) {
				defer wg.Done()
				token, _ := s.generateExchangeToken(stepState.StudentIds[i], entities.UserGroupStudent)
				_, err := bpb.NewLessonModifierServiceClient(s.Conn).Unpublish(s.signedCtxWithToken(ctx, token), &bpb.UnpublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.StudentIds[i]})
				if err != nil {
					reqError <- err
				}
			}(i + idxUnPublish)
		}
	}

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
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to publish, unpublish: %w", Errs)
	}

	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) theNumberOfPublishingStudentsMustBeRecordCorrectly(ctx context.Context) (context.Context, error) {
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return ctx, err

	}

	stepState := StepStateFromContext(ctx)
	expected := stepState.numberOfStudentsArePublishing + stepState.numberOfStudentsWantToPublish - stepState.numberOfStudentsWantToUnpublish
	if numberOfStream != 0 {
		if numberOfStream != expected {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of student, expected: %d, got %d", expected, numberOfStream)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
