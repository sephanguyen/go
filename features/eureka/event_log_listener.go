package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/lestrrat-go/jwx/jwt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) aListOfLearning_objectiveEventLogs(ctx context.Context) (context.Context, error) {
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

func (s *suite) eurekaMustRecordAllStudentsEventLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, ok := stepState.Response.(*epb.CreateStudentEventLogsResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), errors.New("returned data is not CreateStudentEventLogsResponse")
	}
	if !resp.Successful {
		return StepStateToContext(ctx, stepState), errors.New("expected Successful to be true")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfQuiz_finishedEventLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	log := &epb.StudentEventLog{
		EventId:   strconv.Itoa(rand.Int()),
		EventType: "quiz_finished",
		CreatedAt: timestamppb.New(time.Now().Add(-10 * time.Hour)),
		Payload: &epb.StudentEventLogPayload{
			LoId:           stepState.ExistedLoID,
			TimeSpent:      60.0,
			TotalQuestions: 10.0,
			Correct:        7.0,
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
	req.StudentEventLogs = append(req.StudentEventLogs, log)

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfVideo_finishedEventLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	log := &epb.StudentEventLog{
		EventId:   strconv.Itoa(rand.Int()),
		EventType: "video_finished",
		CreatedAt: timestamppb.New(time.Now().Add(-10 * time.Hour)),
		Payload: &epb.StudentEventLogPayload{
			LoId:      stepState.ExistedLoID,
			TimeSpent: 60,
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
	req.StudentEventLogs = append(req.StudentEventLogs, log)

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfStudy_guide_finishedEventLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	log := &epb.StudentEventLog{
		EventId:   strconv.Itoa(rand.Int()),
		EventType: "study_guide_finished",
		CreatedAt: timestamppb.New(time.Now().Add(-10 * time.Hour)),
		Payload: &epb.StudentEventLogPayload{
			LoId:      stepState.ExistedLoID,
			TimeSpent: 60,
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
	req.StudentEventLogs = append(req.StudentEventLogs, log)

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) total_lo_finishedMustBe(ctx context.Context, arg1 string) (context.Context, error) {
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

func (s *suite) total_lo_finishedMustNotBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// sleep to make the evtLogCheckLoop goroutine has a chance to run
	time.Sleep(time.Second)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	row := s.BobDBTrace.QueryRow(ctx, "SELECT COUNT(*) FROM student_statistics WHERE student_id = $1", t.Subject())
	var count int
	if err := row.Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected student statistics is not updated, got: %d", count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aLearningObjectiveIsExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, _ = s.aSignedIn(ctx, "school admin")
	chapters := s.generateChapters(ctx, 1, nil)
	resp, err := epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(s.signedCtx(ctx), &epb.UpsertChaptersRequest{
		Chapters: chapters,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create chapters: %w", err)
	}
	stepState.ChapterID = resp.ChapterIds[0]
	t := s.generateValidTopic(stepState.ChapterID)
	if _, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &epb.UpsertTopicsRequest{
		Topics: []*epb.Topic{&t},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}

	lo := s.generateValidLearningObjectiveEntity(t.Id)

	if _, err := database.Insert(ctx, lo, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.ExistedLoID = lo.ID.String
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) completenessMustBe(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// sleep to make the evtLogCheckLoop goroutine has a chance to run
	time.Sleep(time.Second)

	var scanField interface{}
	switch arg1 {
	case "first_quiz_correctness", "highest_quiz_score":
		scanField = new(pgtype.Float4)
	case "is_finished_quiz", "is_finished_study_guide", "is_finished_video":
		scanField = new(pgtype.Bool)
	}
	query := fmt.Sprintf("SELECT %s FROM students_learning_objectives_completeness WHERE lo_id = $1 ORDER BY created_at DESC", arg1)
	row := s.DB.QueryRow(ctx, query, stepState.ExistedLoID)
	if err := row.Scan(scanField); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	switch v := scanField.(type) {
	case *pgtype.Float4:
		expected, _ := strconv.ParseFloat(arg2, 32)
		if v.Float != float32(expected) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected %s: got %v, want %v", arg1, v.Float, arg2)
		}
	case *pgtype.Bool:
		b, _ := strconv.ParseBool(arg2)
		if v.Bool != b {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected %s: got %v, want %v", arg1, v.Bool, arg2)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInsertsAListOfLearning_objectiveEventLogsThenSleeping(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.aListOfLearning_objectiveEventLogs(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// now stepState.Request is *epb.CreateStudentEventLogsRequest
	resp, err := epb.NewStudentEventLogModifierServiceClient(s.Conn).CreateStudentEventLogs(s.signedCtx(ctx), stepState.Request.(*epb.CreateStudentEventLogsRequest))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if !resp.Successful {
		return StepStateToContext(ctx, stepState), errors.New("couldn't insert a list of event logs")
	}

	duration, err := time.ParseDuration(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	time.Sleep(duration)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) total_learning_timeMustBe(ctx context.Context, arg1 string) (context.Context, error) {
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

func (s *suite) studentInsertsAListOfLearning_objectiveEventLogsWithoutCompletedEventThenSleeping(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SessionID = strconv.Itoa(rand.Int())

	logs := []*epb.StudentEventLog{
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-time.Hour)),
			Payload: &epb.StudentEventLogPayload{
				SessionId: stepState.SessionID,
				Event:     "started",
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-45 * time.Minute)),
			Payload: &epb.StudentEventLogPayload{
				SessionId: stepState.SessionID,
				Event:     "paused",
			},
		},
	}
	req := &epb.CreateStudentEventLogsRequest{StudentEventLogs: logs}

	// now stepState.Request is *epb.CreateStudentEventLogsRequest

	resp, err := epb.NewStudentEventLogModifierServiceClient(s.Conn).CreateStudentEventLogs(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if !resp.Successful {
		return StepStateToContext(ctx, stepState), errors.New("couldn't insert a list of event logs")
	}

	duration, err := time.ParseDuration(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	time.Sleep(duration)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) total_learning_timeMustNotBeExisted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// sleep to make the evtLogCheckLoop goroutine has a chance to run
	time.Sleep(time.Second)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	row := s.BobDBTrace.QueryRow(ctx, "SELECT total_learning_time FROM student_statistics WHERE student_id = $1", t.Subject())
	var count int
	if err := row.Scan(&count); err != nil && err != pgx.ErrNoRows {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) InsertAPresetStudyPlan(ctx context.Context, p *pb.PresetStudyPlan) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &pb.UpsertPresetStudyPlansRequest{
		PresetStudyPlans: []*pb.PresetStudyPlan{p},
	}

	_, err = pb.NewCourseClient(s.BobConn).UpsertPresetStudyPlans(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) UpsertLOs(ctx context.Context, los []*pb.LearningObjective) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	_, err = pb.NewCourseClient(s.Conn).UpsertLOs(s.signedCtx(ctx), &pb.UpsertLOsRequest{
		LearningObjectives: los,
	})
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) studentFinishesAssignedLearningObjectivesOfTheWeek(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var t pgtype.Timestamptz
	switch arg2 {
	case "current":
		t.Set(timeutil.StartWeek())
	case "next":
		t.Set(timeutil.StartDateNextWeek())
	}
	loIDs := stepState.LOIDs

	totalLOs, _ := strconv.ParseInt(arg1, 10, 32)
	if len(loIDs) > int(totalLOs) {
		loIDs = loIDs[0:int(totalLOs)]
	}

	var req *epb.CreateStudentEventLogsRequest
	if stepState.Request == nil {
		req = new(epb.CreateStudentEventLogsRequest)
	} else {
		var ok bool
		req, ok = stepState.Request.(*epb.CreateStudentEventLogsRequest)
		if !ok {
			req = new(epb.CreateStudentEventLogsRequest)
		}
	}
	for _, loID := range loIDs {
		req.StudentEventLogs = append(req.StudentEventLogs, generateFinishedLOEventLogsV1(loID, withFinishQuiz|withFinishVideo|withFinishStudyGuide|withCompletedLO, "")...)
	}
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentFinishesTutorialLo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ChapterID == "" {
		if ctx, err := s.insertAChapter(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve ")
		}
	}
	ctx, _ = s.aSignedIn(ctx, "school admin")
	t := s.generateValidTopic(stepState.ChapterID)
	t.Name = "VN-Tutorial"
	if _, err := epb.NewTopicModifierServiceClient(s.Conn).Upsert(s.signedCtx(ctx), &epb.UpsertTopicsRequest{
		Topics: []*epb.Topic{&t},
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
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

	lo := s.generateValidLearningObjectiveEntity(t.Id)
	lo.ID.Set("LO-VN-Tutorial" + strconv.Itoa(rand.Int()))
	_, err := database.Insert(ctx, lo, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
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

	req.StudentEventLogs = append(req.StudentEventLogs, generateFinishedLOEventLogsV1(lo.ID.String, withFinishQuiz|withFinishVideo|withFinishStudyGuide|withCompletedLO, stepState.StudyPlanItemID)...)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfQuiz_finishedEventLogsWithCorrectnessIs(ctx context.Context, arg1 int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Request == nil {
		if ctx, err := s.aListOfQuiz_finishedEventLogs(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	} else {
		containsQuizFinishedLogs := false
		for _, log := range stepState.Request.(*epb.CreateStudentEventLogsRequest).StudentEventLogs {
			if log.EventType == "quiz_finished" {
				containsQuizFinishedLogs = true
				break
			}
		}
		if !containsQuizFinishedLogs {
			if ctx, err := s.aListOfQuiz_finishedEventLogs(ctx); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	for _, log := range stepState.Request.(*epb.CreateStudentEventLogsRequest).StudentEventLogs {
		if log.EventType == "quiz_finished" {
			totalCorrectQuestions := (float32(arg1) * log.Payload.TotalQuestions) / 100
			log.Payload.Correct = totalCorrectQuestions

			break
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aStudentRetriesTheLastFinishedLearningObjective(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now().UTC()

	stepState.Response, stepState.ResponseErr = epb.NewStudentEventLogModifierServiceClient(s.Conn).CreateStudentEventLogs(s.signedCtx(ctx), stepState.Request.(*epb.CreateStudentEventLogsRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentInsertsAListOfLearning_objectiveEventLogsWithSessionIdEmptyThenSleeping(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.aListOfLearning_objectiveEventLogs(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	logs := stepState.Request.(*epb.CreateStudentEventLogsRequest).StudentEventLogs
	for i := 0; i < len(logs); i++ {
		logs[i].Payload.SessionId = ""
	}

	resp, err := epb.NewStudentEventLogModifierServiceClient(s.Conn).CreateStudentEventLogs(s.signedCtx(ctx), stepState.Request.(*epb.CreateStudentEventLogsRequest))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if !resp.Successful {
		return StepStateToContext(ctx, stepState), errors.New("couldn't insert a list of event logs")
	}

	duration, err := time.ParseDuration(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	time.Sleep(duration)

	return StepStateToContext(ctx, stepState), nil
}
