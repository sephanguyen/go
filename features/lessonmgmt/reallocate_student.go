package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) retrieveStudentsPendingReallocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	req := &lpb.RetrieveStudentPendingReallocateRequest{
		Keyword: "",
		Paging: &cpb.Paging{
			Limit: 1,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
		LessonDate: timestamppb.New(now),
		Filter: &lpb.RetrieveStudentPendingReallocateRequest_Filter{
			StartDate: timestamppb.New(now.AddDate(0, 0, -1)),
			EndDate:   timestamppb.New(now.AddDate(0, 0, 1)),
		},
	}
	stepState.Response, stepState.ResponseErr = lpb.NewStudentSubscriptionServiceClient(s.LessonMgmtConn).RetrieveStudentPendingReallocate(contextWithToken(s, ctx), req)
	stepState.Request2 = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnCorrectReallocateStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	cReq := stepState.Request.(*lpb.CreateLessonRequest)
	res := stepState.Response.(*lpb.RetrieveStudentPendingReallocateResponse)
	for k, v := range res.Items {
		if cReq.StudentInfoList[k].StudentId != v.StudentId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("student_id expected %s,got %s", cReq.StudentInfoList[k].StudentId, v.StudentId)
		}
		if cReq.StudentInfoList[k].CourseId != v.CourseId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("course_id expected %s,got %s", cReq.StudentInfoList[k].CourseId, v.CourseId)
		}
		if stepState.LocationIDs[0] != v.LocationId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("location_id expected %s,got %s", stepState.LocationIDs[0], v.LocationId)
		}
		if stepState.CurrentLessonID != v.OriginalLessonId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("original_lesson_id expected %s,got %s", stepState.CurrentLessonID, v.OriginalLessonId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createLessonWithReallocateStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE)
	for _, st := range req.StudentInfoList {
		st.AttendanceStatus = lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_REALLOCATE
	}
	ctx, err := s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf(err.Error())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) SomeStudentSubscriptions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := stepState.CourseIDs[len(stepState.CourseIDs)-1]
	studentIDWithCourseID := make([]string, 0, len(stepState.StudentIds)*2)
	for _, studentID := range stepState.StudentIds {
		studentIDWithCourseID = append(studentIDWithCourseID, studentID, courseID)
	}
	stepState.StartDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	stepState.EndDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ids, err := s.insertStudentSubscription(ctx, stepState.StartDate, stepState.EndDate, studentIDWithCourseID...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
	}
	stepState.StudentIDWithCourseID = studentIDWithCourseID

	// create access path for above list student subscriptions
	locationId := stepState.LocationIDs[0]
	for _, id := range ids {
		stmt := `INSERT INTO lesson_student_subscription_access_path (student_subscription_id,location_id) VALUES($1,$2)`
		_, err := s.BobDB.Exec(ctx, stmt, id, stepState.LocationIDs[0])
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_student_subscription_access_path with student_subscription_id:%s, location_id:%s, err:%v", id, locationId, err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
