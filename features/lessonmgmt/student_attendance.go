package lessonmgmt

import (
	"context"
	"fmt"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) getStudentAttendance(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.GetStudentAttendanceRequest{
		Paging: &cpb.Paging{
			Limit: 5,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
		Filter: &lpb.GetStudentAttendanceRequest_Filter{},
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewAssignedStudentListServiceClient(s.LessonMgmtConn).GetStudentAttendance(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getStudentAttendanceWithFilter(ctx context.Context, attendanceStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.GetStudentAttendanceRequest{
		Paging: &cpb.Paging{
			Limit: 5,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
		Filter: &lpb.GetStudentAttendanceRequest_Filter{
			AttendanceStatus: []lpb.StudentAttendStatus{lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[attendanceStatus])},
		},
	}

	stepState.FilterAttendanceStatus = attendanceStatus
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewAssignedStudentListServiceClient(s.LessonMgmtConn).GetStudentAttendance(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnCorrectStudentAttendance(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*lpb.GetStudentAttendanceResponse)
	createLessonRequests := stepState.Requests
	createLessonResponses := stepState.Responses
	var totalItems int
	studentCourseMap := map[string]string{}
	for _, req := range createLessonRequests {
		r := req.(*lpb.CreateLessonRequest)
		totalItems += len(r.StudentInfoList)
		for _, s := range r.StudentInfoList {
			studentCourseMap[s.StudentId+s.CourseId] = s.AttendanceStatus.String()
		}
	}
	expectedLessonIDs := []string{}
	for _, res := range createLessonResponses {
		r := res.(*lpb.CreateLessonResponse)
		expectedLessonIDs = append(expectedLessonIDs, r.GetId())
	}
	if len(expectedLessonIDs) != len(resp.GetItems()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("return missing data (expected %d,got %d)", len(expectedLessonIDs), len(resp.GetItems()))
	}
	if len(stepState.FilterAttendanceStatus) > 0 {
		for _, item := range resp.GetItems() {
			if stepState.FilterAttendanceStatus != item.AttendanceStatus.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("return failed data (expected %s,got %s)", stepState.FilterAttendanceStatus, item.AttendanceStatus.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
