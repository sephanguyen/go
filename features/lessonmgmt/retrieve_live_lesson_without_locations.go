package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/lestrrat-go/jwx/jwt"
	"go.uber.org/multierr"
)

func (s *Suite) retrieveLiveLessonByCourseWithStartTimeAndEndTime(ctx context.Context, userRole, courseID, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	courseID += stepState.Random

	userEntity := ""
	if userRole == "student" {
		userEntity = entities.UserGroupStudent
	}

	if userRole == "teacher" {
		userEntity = entities.UserGroupTeacher
	}
	token, err := s.CommonSuite.GenerateExchangeToken(stepState.CurrentTeacherID, userEntity)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		CourseIds: []string{courseID},
		From:      &types.Timestamp{Seconds: startDate.Unix()},
		To:        &types.Timestamp{Seconds: endDate.Unix()},
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.BobConn).RetrieveLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobMustReturnPbLiveLessonForStudent(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if result == "empty" {
		return s.bobMustReturnEmptyLiveLesson(ctx)
	}
	if result == "correct" {
		return s.bobMustReturnCorrectLiveLessonsForStudent(ctx)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobReturnResultLiveLessonForTeacher(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if result == "empty" {
		return s.bobMustReturnEmptyLiveLesson(ctx)
	}
	if result == "correct" {
		return s.bobMustReturnCorrectLiveLessonsForTeacher(ctx)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobMustReturnEmptyLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	if len(rsp.Lessons) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return empty lessons")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobMustReturnCorrectLiveLessonsForStudent(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.bobMustReturnLessons(ctx)
	ctx, err2 := s.returnPbLessonsMustHaveCorrectStatus(ctx)
	ctx, err3 := s.returnPbLessonsMustHaveCorrectTeacherProfile(ctx)
	ctx, err4 := s.returnCorrectUserClassIdInLesson(ctx)
	ctx, err5 := s.resultCorrectPbLessonMember(ctx)
	ctx, err6 := s.returnPbLessonsMustHaveCorrectSchedulingStatus(ctx)
	err := multierr.Combine(err1, err2, err3, err4, err5, err6)
	return ctx, err
}

func (s *Suite) bobMustReturnCorrectLiveLessonsForTeacher(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.bobMustReturnLessons(ctx)
	ctx, err2 := s.returnPbLessonsMustHaveCorrectStatus(ctx)
	ctx, err3 := s.returnPbLessonsMustHaveCorrectTeacherProfile(ctx)
	ctx, err4 := s.returnCorrectUserClassIdInLesson(ctx)
	ctx, err5 := s.resultCorrectPbLessonCourse(ctx)
	ctx, err6 := s.returnPbLessonsMustHaveCorrectSchedulingStatus(ctx)
	err := multierr.Combine(err1, err2, err3, err4, err5, err6)
	return ctx, err
}

func (s *Suite) returnPbLessonsMustHaveCorrectSchedulingStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	lessonRepo := repo.LessonRepo{}
	for _, lesson := range rsp.Lessons {
		lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, lesson.LessonId)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("get lesson by lesson Id %s fail: %w", lesson.LessonID, err)
		}
		if lesson.SchedulingStatus == domain.LessonSchedulingStatusPublished {
			continue
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s return wrong status: %s", lesson.LessonID, lesson.SchedulingStatus)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnPbLessonsMustHaveCorrectStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	for _, lesson := range rsp.Lessons {
		status := lesson.Status
		if lesson.EndTime.Nanos < int32(time.Now().Nanosecond()) && status == pb.LESSON_STATUS_COMPLETED {
			continue
		}
		if lesson.StartTime.Nanos >= int32(time.Now().Nanosecond()) && status == pb.LESSON_STATUS_IN_PROGRESS {
			continue
		}
		if lesson.StartTime.Nanos < int32(time.Now().Nanosecond()) && status == pb.LESSON_STATUS_NOT_STARTED {
			continue
		}
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s return wrong status: %s", lesson.LessonId, status.String())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnCorrectUserClassIdInLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	courseMapIDs := make(map[string]bool)
	for _, lesson := range rsp.Lessons {
		found, ok := courseMapIDs[lesson.CourseId]
		if found && ok {
			continue
		}
		courseMapIDs[lesson.CourseId] = true
		ctx, err := s.checkUserClassIDInLesson(ctx, lesson)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectPbLessonMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	lessonIDs := []string{}
	for _, lesson := range rsp.Lessons {
		lessonIDs = append(lessonIDs, lesson.LessonId)
	}
	t, _ := jwt.ParseString(stepState.AuthToken)
	query := `SELECT result_lesson
	FROM UNNEST($1::TEXT[]) AS result_lesson
	LEFT JOIN lesson_members lm ON result_lesson=lm.lesson_id AND lm.user_id =$2
	WHERE lm.lesson_id IS NULL`
	rows, err := s.BobDB.Query(ctx, query, database.TextArray(lessonIDs), t.Subject())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	wrongLessons := []string{}
	for rows.Next() {
		var wrongLesson string
		if err := rows.Scan(&wrongLesson); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("rows.Scan :%v", err)
		}
		wrongLessons = append(wrongLessons, wrongLesson)
	}
	if len(wrongLessons) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student not a member of these lesson %s", wrongLessons)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) resultCorrectPbLessonCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.RetrieveLiveLessonRequest)
	if len(req.CourseIds) == 0 {
		return StepStateToContext(ctx, stepState), nil
	}
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	for _, lesson := range rsp.GetLessons() {
		if !contains(req.CourseIds, lesson.CourseId) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s not match course retrieve\n%s\n%s", lesson.LessonId, req.CourseIds, lesson.CourseId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnPbLessonsMustHaveCorrectTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	for _, lesson := range rsp.Lessons {
		if len(lesson.Teacher) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson return does not contain teacher")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
