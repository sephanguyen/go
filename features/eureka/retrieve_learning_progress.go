package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aSignedInStudentWithFilterRangeIs(ctx context.Context, filterRange string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedIn(ctx, constant.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	switch filterRange {
	case "valid":
		ctx, err = s.aValidFilterRange(ctx)
	case "invalid":
		ctx, err = s.anInvalidFilterRange(ctx)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInStudentWithFilterRangeIsUseHisOwnedStudentUUID(ctx context.Context, filterRange string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedInStudentWithFilterRangeIs(ctx, filterRange)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.hisOwnedStudentUUID(ctx)

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aSignedInStudentWithUseHisOwnedStudentUUID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedIn(ctx, "student")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.hisOwnedStudentUUID(ctx)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anInvalidFilterRange(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.From = timestamppb.Now()
	stepState.To = timestamppb.New(time.Now().Add(-1 * time.Minute))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidFilterRange(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.From = timestamppb.New(timeutil.StartWeekIn(bpb.Country(cpb.Country_COUNTRY_VN)))
	stepState.To = timestamppb.New(timeutil.EndWeekIn(bpb.Country(cpb.Country_COUNTRY_VN)))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aFilterRangeWithIs(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	t, err := time.Parse(time.RFC3339, arg2)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	switch arg1 {
	case "from":
		stepState.From = timestamppb.New(t)
	case "to":
		stepState.To = timestamppb.New(t)
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unknown %q", arg1)
	}
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) studentHasNotLearnedAnyLO(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SessionID = "any session that doesn't exist"
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentRetrievesLP(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(time.Second)

	var studentID string
	switch {
	case len(stepState.OtherStudentIDs) > 0:
		studentID = stepState.OtherStudentIDs[0]
	case stepState.CurrentStudentID != "":
		studentID = stepState.CurrentStudentID
	}
	stepState.Response, stepState.ResponseErr = pb.NewStudentLearningTimeReaderClient(s.Conn).RetrieveLearningProgress(s.signedCtx(ctx), &pb.RetrieveLearningProgressRequest{
		StudentId: studentID,
		SessionId: stepState.SessionID,
		From:      stepState.From,
		To:        stepState.To,
	})

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsLPWithSomeTotal_time_spent_in_dayLargerThanZero(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveLearningProgressResponse)

	someTimeSpentNotEqualsZero := false
	for _, d := range resp.Dailies {
		if d.TotalTimeSpentInDay > 0 {
			someTimeSpentNotEqualsZero = true
		}
	}
	if !someTimeSpentNotEqualsZero {
		return StepStateToContext(ctx, stepState), errors.New("expected some time spent not equal to zero")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsLPWithAllTotal_time_spent_in_dayEqualToZero(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveLearningProgressResponse)

	allTimeSpentEqualZero := true
	for _, d := range resp.Dailies {
		if d.TotalTimeSpentInDay > 0 {
			allTimeSpentEqualZero = false
			break
		}
	}
	if !allTimeSpentEqualZero {
		return StepStateToContext(ctx, stepState), errors.New("expected all total_time_spent_in_day equal to zero")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aLearningObjectiveEventLogWithSessionAt(ctx context.Context, event, sessionID, createdAt string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	t, _ := time.Parse(time.RFC3339, createdAt)
	ctx, studyPlanItems, err := s.generateAStudyPlanItemV2(ctx)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	s.aSignedIn(ctx, "school admin")
	stepState.StudyPlanItems = studyPlanItems
	_, err = epb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlanItemV2(s.signedCtx(ctx), &epb.UpsertStudyPlanItemV2Request{
		StudyPlanItems: studyPlanItems,
	})
	if err != nil {
		fmt.Printf("can not upsert study plan item v2:%v\n", err)
	}
	stepState.AuthToken = stepState.StudentToken
	s.signedCtx(ctx)

	logs := []*epb.StudentEventLog{
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(t),
			Payload: &pb.StudentEventLogPayload{
				SessionId:       sessionID + stepState.Random,
				Event:           event,
				LoId:            "lo_id",
				StudyPlanItemId: stepState.StudyPlanItemID,
			},
		},
	}

	var req *epb.CreateStudentEventLogsRequest
	if stepState.Request == nil {
		req = new(epb.CreateStudentEventLogsRequest)
	} else {
		var ok bool
		req, ok = stepState.Request.(*epb.CreateStudentEventLogsRequest)
		if !ok {
			return StepStateToContext(ctx, stepState), errors.New("stepState.Request should be *epb.CreateStudentEventLogsRequest")
		}
	}
	req.StudentEventLogs = append(req.StudentEventLogs, logs...)

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) total_learning_timeAtMustBe(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	day, err := time.Parse(time.RFC3339, arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	learningTime, err := strconv.Atoi(arg2)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, d := range stepState.Response.(*epb.RetrieveLearningProgressResponse).Dailies {
		t := d.Day.AsTime()
		if t.Equal(day) {
			if d.TotalTimeSpentInDay == int64(learningTime) {
				return StepStateToContext(ctx, stepState), nil
			} else {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total learning time in: %q, got: %v, expected: %v", day, d.TotalTimeSpentInDay, learningTime)
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) previousRequestDataIsReset(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = nil
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherRetrievesLP(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	time.Sleep(time.Second)
	studentID := idutil.ULIDNow()
	ctx, err := s.aValidUser(ctx, studentID, constant.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidUser: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewStudentLearningTimeReaderClient(s.Conn).RetrieveLearningProgress(s.signedCtx(ctx), &pb.RetrieveLearningProgressRequest{
		StudentId: studentID,
		SessionId: stepState.SessionID,
		From:      stepState.From,
		To:        stepState.To,
	})

	return StepStateToContext(ctx, stepState), nil
}
