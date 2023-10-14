package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"golang.org/x/exp/slices"
)

func (s *Suite) userGetListStudentSubscriptionsInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &lpb.GetStudentCourseSubscriptionsRequest{
		Subscriptions: make([]*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription, 0, len(stepState.StudentIDWithCourseID)),
	}
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		req.Subscriptions = append(req.Subscriptions, &lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
			StudentId: stepState.StudentIDWithCourseID[i],
			CourseId:  stepState.StudentIDWithCourseID[i+1],
		})
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = lpb.NewStudentSubscriptionServiceClient(s.LessonMgmtConn).
		GetStudentCourseSubscriptions(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) gotListStudentSubscriptionsInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*lpb.GetStudentCourseSubscriptionsResponse)
	req := stepState.Request.(*lpb.GetStudentCourseSubscriptionsRequest)

	if len(res.Items) != len(req.Subscriptions) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d items but got %d", len(req.Subscriptions), len(res.Items))
	}

	actual := make(map[string]*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription)
	for i, item := range res.Items {
		actual[item.StudentId] = res.Items[i]
	}

	for _, expected := range req.Subscriptions {
		v, ok := actual[expected.StudentId]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("could not found student id %s", expected.StudentId)
		}
		if v.CourseId != expected.CourseId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected course id %s but got %s", expected.CourseId, v.CourseId)
		}
		if !stringutil.SliceElementsMatch(stepState.LocationIDs, v.LocationIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected location id %v of student subscription %s but got %v", stepState.LocationIDs, v.Id, v.LocationIds)
		}
		if !v.GetStartDate().AsTime().Equal(stepState.StartDate) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected start at %s, but got %s", stepState.StartDate.String(), v.GetStartDate().AsTime().String())
		}
		if !v.GetEndDate().AsTime().Equal(stepState.EndDate) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected end at %s, but got %s", stepState.EndDate.String(), v.GetEndDate().AsTime().String())
		}
		fmt.Println("grade v2", v.GetGrade(), v.GetGradeV2())
		if !slices.Contains(stepState.GradeIDs, v.GetGradeV2()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected grade at %s, but got %s", stepState.GradeIDs, v.GetGradeV2())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
