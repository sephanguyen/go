package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) userRetrievesStudentStats(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(1000 * time.Millisecond)

	var studentID string
	switch {
	case len(stepState.OtherStudentIDs) > 0:
		studentID = stepState.OtherStudentIDs[0]
	case len(stepState.UnAssignedStudentIDs) > 0:
		studentID = stepState.UnAssignedStudentIDs[0]
	case len(stepState.AssignedStudentIDs) > 0:
		studentID = stepState.AssignedStudentIDs[0]
	case stepState.CurrentStudentID != "":
		studentID = stepState.CurrentStudentID
	}
	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanReaderServiceClient(s.Conn).RetrieveStat(s.signedCtx(ctx), &epb.RetrieveStatRequest{
		StudentId: studentID,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) achievementCrownMustBe(ctx context.Context, crownArg string, totalArg int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// sleep to make the evtLogCheckLoop goroutine has a chance to run
	time.Sleep(time.Second)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, err := epb.NewStudyPlanReaderServiceClient(s.Conn).RetrieveStat(s.signedCtx(ctx), &epb.RetrieveStatRequest{
		StudentId: t.Subject(),
	})

	stat := resp.StudentStat

	var total int32 = -1
	for _, crown := range stat.Crowns {
		if crown.AchievementCrown == crownArg {
			total = crown.Total
			break
		}
	}

	if total != int32(totalArg) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total for crown %q, got %d, want %d", crownArg, total, totalArg)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) totalLoFinishedMustBe(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// sleep to make the evtLogCheckLoop goroutine has a chance to run
	time.Sleep(time.Second)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, err := epb.NewStudyPlanReaderServiceClient(s.Conn).RetrieveStat(s.signedCtx(ctx), &epb.RetrieveStatRequest{
		StudentId: t.Subject(),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expected, _ := strconv.ParseInt(arg1, 10, 32)
	if resp.StudentStat.TotalLoFinished != int32(expected) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected total_lo_finished: %d, got: %d", expected, resp.StudentStat.TotalLoFinished)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) totalLearningTimeMustBe(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, err := epb.NewStudyPlanReaderServiceClient(s.Conn).RetrieveStat(s.signedCtx(ctx), &epb.RetrieveStatRequest{
		StudentId: t.Subject(),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expected, err := time.ParseDuration(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if resp.StudentStat.TotalLearningTime != int32(expected/time.Second) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected total_learning_time: %v, got: %d", expected, resp.StudentStat.TotalLearningTime)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) totalLearningTimeMustBeV2(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, err := epb.NewStudyPlanReaderServiceClient(s.Conn).RetrieveStatV2(s.signedCtx(ctx), &epb.RetrieveStatRequest{
		StudentId: t.Subject(),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	expected, err := time.ParseDuration(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if resp.StudentStat.TotalLearningTime != int32(expected/time.Second) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected total_learning_time: %v, got: %d", expected, resp.StudentStat.TotalLearningTime)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfLearningObjectiveEventLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SessionID = strconv.Itoa(rand.Int())
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
			CreatedAt: timestamppb.New(time.Now().Add(-time.Hour)),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       stepState.SessionID,
				Event:           "started",
				LoId:            "lo_id",
				StudyPlanItemId: stepState.StudyPlanItemID,
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-45 * time.Minute)),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       stepState.SessionID,
				Event:           "paused",
				LoId:            "lo_id",
				StudyPlanItemId: stepState.StudyPlanItemID,
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-15 * time.Minute)),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       stepState.SessionID,
				Event:           "resumed",
				LoId:            "lo_id",
				StudyPlanItemId: stepState.StudyPlanItemID,
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now()),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       stepState.SessionID,
				Event:           "completed",
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

func (s *suite) aStudentInsertsAListOfEventLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now().UTC()
	stepState.AuthToken = stepState.StudentToken

	stepState.Response, stepState.ResponseErr = epb.NewStudentEventLogModifierServiceClient(s.Conn).CreateStudentEventLogs(s.signedCtx(ctx), stepState.Request.(*epb.CreateStudentEventLogsRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hisOwnedStudentUUID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return ctx, err
	}
	stepState.CurrentStudentID = t.Subject()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anOtherStudentProfileInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	orgtoken := stepState.AuthToken
	if ctx, err := s.aValidStudentInDB(ctx, id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = orgtoken
	stepState.OtherStudentIDs = append(stepState.OtherStudentIDs, id)
	stepState.UserID = id

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentRetrievesPresetStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &pb.RetrievePresetStudyPlansRequest{
		Grade:   "G12",
		Country: pb.COUNTRY_VN,
	}
	stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).RetrievePresetStudyPlans(s.signedCtx(ctx), stepState.Request.(*pb.RetrievePresetStudyPlansRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentFinishesUnassignedLearningObjectives(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, _ = s.aSignedIn(ctx, "school admin")
	bookResp, err := epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(s.signedCtx(ctx), &epb.UpsertBooksRequest{
		Books: s.generateBooks(1, nil),
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]

	stepState.SchoolIDInt = constants.ManabieSchool
	if stepState.ChapterID == "" {
		resp, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &epb.UpsertChaptersRequest{
			Chapters: s.generateChapters(ctx, 1, nil),
			BookId:   stepState.BookID,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
		}
		stepState.ChapterID = resp.ChapterIds[0]
	}

	t := s.generateValidTopic(stepState.ChapterID)
	t.Name = "Topic-G12-Math-1"

	stepState.OldToken = stepState.AuthToken
	stepState.AuthToken = stepState.SchoolAdminToken
	resp, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(
		s.signedCtx(ctx), &epb.UpsertTopicsRequest{
			Topics: []*epb.Topic{&t},
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TopicID = resp.GetTopicIds()[0]
	stepState.AuthToken = stepState.OldToken
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

	ctx, studyPlanItems, err := s.generateAStudyPlanItemV2(ctx)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	stepState.StudyPlanItems = studyPlanItems
	_, err = epb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlanItemV2(s.signedCtx(ctx), &epb.UpsertStudyPlanItemV2Request{
		StudyPlanItems: studyPlanItems,
	})
	if err != nil {
		fmt.Printf("can not upsert study plan item v2:%v\n", err)
	}
	stepState.AuthToken = stepState.StudentToken
	s.signedCtx(ctx)
	totalLOs, _ := strconv.ParseInt(arg1, 10, 32)
	los := make([]*cpb.LearningObjective, 0, totalLOs)
	for i := 0; i < int(totalLOs); i++ {
		lo := s.generateLearningObjective1(ctx)
		los = append(los, lo)

		req.StudentEventLogs = append(req.StudentEventLogs, generateFinishedLOEventLogsV1(lo.Info.Id, withFinishQuiz|withFinishVideo|withFinishStudyGuide|withCompletedLO, stepState.StudyPlanItemID)...)
	}
	stepState.OldToken = stepState.AuthToken
	stepState.AuthToken = stepState.SchoolAdminToken

	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(s.signedCtx(ctx), &epb.UpsertLOsRequest{
		LearningObjectives: los,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = stepState.OldToken
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) InsertATopic(ctx context.Context, p *epb.Topic) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, _ = s.aSignedIn(ctx, "school admin")

	_, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &epb.UpsertTopicsRequest{
		Topics: []*epb.Topic{p},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

const (
	withFinishQuiz = 1 << iota
	withFinishVideo
	withFinishStudyGuide
	withCompletedLO
)

func generateFinishedLOEventLogs(loID string, flag int) (logs []*pb.StudentEventLog) {
	if flag&withFinishQuiz != 0 {
		logs = append(logs, &pb.StudentEventLog{EventId: strconv.Itoa(rand.Int()), EventType: "quiz_finished", CreatedAt: &types.Timestamp{Seconds: time.Now().Add(-10 * time.Hour).Unix()}, Payload: &types.Struct{Fields: map[string]*types.Value{"lo_id": {Kind: &types.Value_StringValue{StringValue: loID}}, "time_spent": {Kind: &types.Value_NumberValue{NumberValue: 60.0}}, "total_questions": {Kind: &types.Value_NumberValue{NumberValue: 10.0}}, "correct": {Kind: &types.Value_NumberValue{NumberValue: 7.0}}}}})
	}
	if flag&withFinishVideo != 0 {
		logs = append(logs, &pb.StudentEventLog{EventId: strconv.Itoa(rand.Int()), EventType: "video_finished", CreatedAt: &types.Timestamp{Seconds: time.Now().Add(-10 * time.Hour).Unix()}, Payload: &types.Struct{Fields: map[string]*types.Value{"lo_id": {Kind: &types.Value_StringValue{StringValue: loID}}, "time_spent": {Kind: &types.Value_NumberValue{NumberValue: 60.0}}}}})
	}
	if flag&withFinishStudyGuide != 0 {
		logs = append(logs, &pb.StudentEventLog{EventId: strconv.Itoa(rand.Int()), EventType: "study_guide_finished", CreatedAt: &types.Timestamp{Seconds: time.Now().Add(-10 * time.Hour).Unix()}, Payload: &types.Struct{Fields: map[string]*types.Value{"lo_id": {Kind: &types.Value_StringValue{StringValue: loID}}, "time_spent": {Kind: &types.Value_NumberValue{NumberValue: 60.0}}}}})
	}
	if flag&withCompletedLO != 0 {
		logs = append(logs, &pb.StudentEventLog{EventId: strconv.Itoa(rand.Int()), EventType: "learning_objective", CreatedAt: &types.Timestamp{Seconds: time.Now().Add(-10 * time.Hour).Unix()}, Payload: &types.Struct{Fields: map[string]*types.Value{"event": {Kind: &types.Value_StringValue{StringValue: "completed"}}, "lo_id": {Kind: &types.Value_StringValue{StringValue: loID}}}}})
	}
	return
}

func generateFinishedLOEventLogsV1(loID string, flag int, studyPlanItemID string) (logs []*epb.StudentEventLog) {
	if flag&withFinishQuiz != 0 {
		logs = append(logs, &epb.StudentEventLog{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "quiz_finished",
			CreatedAt: timestamppb.New(time.Now().Add(-10 * time.Hour)),
			Payload: &epb.StudentEventLogPayload{
				LoId:            loID,
				TimeSpent:       60.0,
				TotalQuestions:  10.0,
				Correct:         7.0,
				StudyPlanItemId: studyPlanItemID,
			},
		})
	}
	if flag&withFinishVideo != 0 {
		logs = append(logs, &epb.StudentEventLog{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "video_finished",
			CreatedAt: timestamppb.New(time.Now().Add(-10 * time.Hour)),
			Payload: &epb.StudentEventLogPayload{
				LoId:            loID,
				TimeSpent:       60.0,
				StudyPlanItemId: studyPlanItemID,
			},
		})
	}
	if flag&withFinishStudyGuide != 0 {
		logs = append(logs, &epb.StudentEventLog{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "study_guide_finished",
			CreatedAt: timestamppb.New(time.Now().Add(-10 * time.Hour)),
			Payload: &epb.StudentEventLogPayload{
				LoId:            loID,
				TimeSpent:       60.0,
				StudyPlanItemId: studyPlanItemID,
			},
		})
	}
	if flag&withCompletedLO != 0 {
		logs = append(logs, &epb.StudentEventLog{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-10 * time.Hour)),
			Payload: &epb.StudentEventLogPayload{
				Event:           "completed",
				LoId:            loID,
				StudyPlanItemId: studyPlanItemID,
			},
		})
	}
	return
}
