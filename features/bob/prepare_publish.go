package bob

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/status"
)

func (s *suite) someValidLearnersInDatabase(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.totalNumberOfStudents = rand.Intn(5) + s.Cfg.Agora.MaximumLearnerStreamings // range [MaximumLearnerStreamings-20]
	stepState.StudentIds = make([]string, 0, stepState.totalNumberOfStudents)
	for i := 0; i < stepState.totalNumberOfStudents; i++ {
		id := s.newID()
		if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.StudentIds = append(stepState.StudentIds, id)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidLessonInDB(ctx context.Context) (context.Context, error) {
	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
	stepState := StepStateFromContext(ctx)

	e := &entities.Lesson{}
	now := timeutil.Now()
	e.LessonID.Set(idutil.ULID(now))
	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)
	stmt := `INSERT INTO lessons(lesson_id,created_at,updated_at,center_id) VALUES($1,$2,$3,$4)`
	_, err := s.DB.Exec(ctx, stmt, e.LessonID, e.CreatedAt, e.UpdatedAt, constants.ManabieOrgLocation)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidLessonInDB: %w", err)
	}
	stepState.lessonID = e.LessonID.String
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aNumberOfStreamOfTheLesson(ctx context.Context, args string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var numberOfStream int
	if args == "maximum" {
		numberOfStream = s.Cfg.Agora.MaximumLearnerStreamings
	} else if args == "second maximum" {
		numberOfStream = s.Cfg.Agora.MaximumLearnerStreamings - 1
	}
	stmt := `UPDATE lessons SET stream_learner_counter=$1 WHERE lesson_id=$2`
	_, err := s.DB.Exec(ctx, stmt, numberOfStream, stepState.lessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aDefaultNumberOfStreamOfTheLesson: %w", err)
	}
	stepState.numberOfStream = numberOfStream
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLearnerPrepareToPublish(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.AuthToken, err = s.generateExchangeToken(stepState.studentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(s.signedCtx(ctx), &bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.studentID})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkLearnerIsPublish(ctx context.Context, learnerID string) (context.Context, bool, error) {
	stepState := StepStateFromContext(ctx)
	ID := stepState.studentID
	if learnerID != "" {
		ID = learnerID
	}
	var isPushlishing bool
	checkLearnerIsPublishStmt := `SELECT EXISTS(SELECT * FROM lessons WHERE lesson_id=$1 AND $2 =ANY(learner_ids))`
	err := s.DB.QueryRow(ctx, checkLearnerIsPublishStmt, stepState.lessonID, ID).Scan(&isPushlishing)
	return ctx, isPushlishing, err
}
func (s *suite) newRecordIndicatingThatTheLearnerIsPublishingAnUploadStreamInTheLesson(ctx context.Context, args string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var isPushlishing bool
	var err error
	switch args {
	case "first":
		ctx, isPushlishing, err = s.checkLearnerIsPublish(ctx, stepState.firstLearner)
	case "second":
		ctx, isPushlishing, err = s.checkLearnerIsPublish(ctx, stepState.secondLearner)
	default:
		ctx, isPushlishing, err = s.checkLearnerIsPublish(ctx, "")

	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if !isPushlishing {
		return StepStateToContext(ctx, stepState), fmt.Errorf("none publishing streaming")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) removelearnerStream(ctx context.Context, lessonID, learnerID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonStreamRepo := &repositories.LessonRepo{}
	if err := lessonStreamRepo.DecreaseNumberOfStreaming(ctx, s.DB, database.Text(lessonID), database.Text(learnerID)); err != nil {
		// if no row effected ->it's ok because we just make sure the learner's not publishing stream
		if err == repositories.ErrUnAffected {
			return StepStateToContext(ctx, stepState), nil
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("removelearnerStream: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theAbitraryLearnerDoesNotPublishingAnyUploadingStreamInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.removelearnerStream(ctx, stepState.lessonID, stepState.studentID)
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) getNumberOfStreamOfTheLesson(ctx context.Context) (context.Context, int, error) {
	stmt := `SELECT stream_learner_counter FROM lessons WHERE lesson_id=$1`
	var numberOfStream int
	stepState := StepStateFromContext(ctx)
	err := s.DB.QueryRow(ctx, stmt, stepState.lessonID).Scan(&numberOfStream)
	return ctx, numberOfStream, err
}
func (s *suite) theNumberOfStreamOfTheLessonHaveToIncreasing(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if numberOfStream <= stepState.numberOfStream {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of stream haven't changed")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsPreparePublishTheResponse(ctx context.Context, args string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	resp, ok := (stepState.Response).(*bpb.PreparePublishResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Response must be *pb.PreparePublishResponse")
	}
	if resp.Status.String() != args {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected status of response: expected :%s, got: %s", args, resp.Status)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLearnerIsNotAllowedToPublishAnyUploadingStreamInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, isPublish, err := s.checkLearnerIsPublish(ctx, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if isPublish {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the learner not allowed to publish any uploading stream in the lesson")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLearnerIsStillPublishingAnUploadingStreamInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, isPublish, err := s.checkLearnerIsPublish(ctx, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if !isPublish {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the learner have to still publishing the uploading stream")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theNumberOfStreamOfTheLessonHaveToNoChange(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if numberOfStream != stepState.numberOfStream {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the number of the lesson have to no change")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) anArbitraryLearnerPublishingAnUploadingStreamInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.AuthToken, err = s.generateExchangeToken(stepState.studentID, entities.UserGroupAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	_, err = bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(s.signedCtx(ctx), &bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.studentID})
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	// plus 1 for current
	stepState.numberOfStream += 1
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLearnerCurrentlyDoesNotPublishAnyUploadingStreamInTheLesson(ctx context.Context, args string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch args {
	case "first":
		stepState.firstLearner = stepState.StudentIds[0]
		if ctx, err := s.removelearnerStream(ctx, stepState.lessonID, stepState.firstLearner); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "second":
		stepState.secondLearner = stepState.StudentIds[1]
		if ctx, err := s.removelearnerStream(ctx, stepState.lessonID, stepState.secondLearner); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) signedCtxWithToken(ctx context.Context, token string) context.Context {
	return helper.GRPCContext(ctx, "token", token)
}
func (s *suite) twoLearnersPrepareToPublishInConcurrently(ctx context.Context) (context.Context, error) {
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
		stepState.firstResponse, stepState.firstResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(
			s.signedCtxWithToken(ctx, firstToken), &bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.firstLearner})
	}()
	go func() {
		defer wg.Done()
		var err error
		secondToken, err := s.generateExchangeToken(stepState.secondLearner, entities.UserGroupStudent)
		if err != nil {
			stepState.secondResponseErr = err
			return
		}
		stepState.secondResponse, stepState.secondResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(
			s.signedCtxWithToken(ctx, secondToken),
			&bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.secondLearner})
	}()
	wg.Wait()
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLessonsLearnerCounterHaveToIncreasingTwo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if numberOfStream != stepState.numberOfStream+2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson's learner counter have to increasing two")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsStatusForBothRequests(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stt1, ok := status.FromError(stepState.firstResponseErr)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error in first request is not status.Status, err: %s", stepState.firstResponseErr.Error())
	}
	stt2, ok := status.FromError(stepState.secondResponseErr)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error in second request is not status.Status, err: %s", stepState.secondResponseErr.Error())
	}
	if stt1.Code() != stt2.Code() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("status code of 2 request have to equal")
	}
	if stt1.Code().String() != arg1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting req1 %s, got %s status code, message: %s", arg1, stt1.Code().String(), stt1.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) newRecordIndicatingThatOneOfThemPublishing(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, isPublish1, err := s.checkLearnerIsPublish(ctx, stepState.firstLearner)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.checkLearnerIsPublish-firstlearner: %w", err)
	}
	ctx, isPublish2, err := s.checkLearnerIsPublish(ctx, stepState.secondLearner)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.checkLearnerIsPublish-secondlearner: %w", err)
	}
	if !isPublish1 && !isPublish2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("one of two request have to publish success")
	}
	if isPublish1 == isPublish2 && isPublish1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("two request cannot publish success concurrently")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLessonsLearnerCounterHaveToMaximum(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	if numberOfStream != s.Cfg.Agora.MaximumLearnerStreamings {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson's learner counter have to equal maximum")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsForTheUserWhoIsGranted(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stt1, ok := status.FromError(stepState.firstResponseErr)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error in first request is not status.Status, err: %s", stepState.firstResponseErr.Error())
	}
	stt2, ok := status.FromError(stepState.secondResponseErr)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error in second request is not status.Status, err: %s", stepState.secondResponseErr.Error())
	}
	if stt1.Code() != stt2.Code() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("status code of 2 request have to equal")
	}
	if stt1.Code().String() != arg1 || stt2.Code().String() != arg1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s, got  req1 %s, req2 %s,  status code", arg1, stt1.Code().String(), stt2.Code())
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsTheResponseForTheUserWhoIsRejected(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.ResponseErr = stepState.firstResponseErr
	stepState.Response = stepState.firstResponse
	ctx, err1 := s.returnsPreparePublishTheResponse(ctx, arg1)
	stepState.ResponseErr = stepState.secondResponseErr
	stepState.Response = stepState.secondLearner
	ctx, err2 := s.returnsPreparePublishTheResponse(ctx, arg1)
	if err1 == nil && err2 == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("both request have to failed one")
	}
	if err1 == nil && err2 == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("both request cannot success both")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theLearnersPrepareToPublishTwiceInConcurrently(ctx context.Context) (context.Context, error) {
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

		stepState.firstResponse, stepState.firstResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(
			s.signedCtxWithToken(ctx, token),
			&bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.studentID})
	}()
	go func() {
		defer wg.Done()
		var err error
		token, err := s.generateExchangeToken(stepState.studentID, entities.UserGroupStudent)
		if err != nil {
			stepState.secondResponseErr = err
			return
		}

		stepState.secondResponse, stepState.secondResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).PreparePublish(
			s.signedCtxWithToken(ctx, token),
			&bpb.PreparePublishRequest{LessonId: stepState.lessonID, LearnerId: stepState.studentID})
	}()
	wg.Wait()
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theNumberOfStreamOfTheLessonHaveToBecome(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedNumberOfStream, err := strconv.ParseInt(arg1, 10, 64)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}
	ctx, numberOfStream, err := s.getNumberOfStreamOfTheLesson(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if numberOfStream != int(expectedNumberOfStream) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of stream, expected %d, got %d", expectedNumberOfStream, numberOfStream)
	}
	return StepStateToContext(ctx, stepState), nil
}
