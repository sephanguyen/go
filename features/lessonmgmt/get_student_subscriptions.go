package lessonmgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func (s *Suite) userGetListStudentSubscriptions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.GetStudentCourseSubscriptionsRequest{
		Subscriptions: make([]*bpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription, 0, len(stepState.StudentIDWithCourseID)),
	}
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		req.Subscriptions = append(req.Subscriptions, &bpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
			StudentId: stepState.StudentIDWithCourseID[i],
			CourseId:  stepState.StudentIDWithCourseID[i+1],
		})
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewStudentSubscriptionServiceClient(s.Connections.BobConn).
		GetStudentCourseSubscriptions(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) gotListStudentSubscriptions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res := stepState.Response.(*bpb.GetStudentCourseSubscriptionsResponse)
	req := stepState.Request.(*bpb.GetStudentCourseSubscriptionsRequest)

	if len(res.Items) != len(req.Subscriptions) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d items but got %d", len(req.Subscriptions), len(res.Items))
	}

	actual := make(map[string]*bpb.GetStudentCourseSubscriptionsResponse_StudentSubscription)
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
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserRetrieveStudentSubscription(ctx context.Context, limit, offset int, keyword, coursers, grades string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	filter := &bpb.RetrieveStudentSubscriptionFilter{}
	keywordStr := ""
	if keyword != "" {
		keywordStr = keyword
	}
	if coursers != "" && len(strings.Split(coursers, ",")) > 0 {
		filter.CourseId = stepState.FilterCourseIDs[0:len(strings.Split(coursers, ","))]
	}
	if len(grades) > 0 {
		filter.Grade = strings.Split(grades, ",")
	}
	req := &bpb.RetrieveStudentSubscriptionRequest{
		Paging: &cpb.Paging{
			Limit: uint32(limit),
		},
		Keyword: keywordStr,
		Filter:  filter,
	}

	if offset > 0 {
		offset := stepState.FilterStudentSubs[offset]
		req = &bpb.RetrieveStudentSubscriptionRequest{
			Paging: &cpb.Paging{
				Limit:  uint32(limit),
				Offset: &cpb.Paging_OffsetString{OffsetString: offset},
			},
			Keyword: keywordStr,
			Filter:  filter,
		}
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = bpb.NewStudentSubscriptionServiceClient(s.Connections.BobConn).
		RetrieveStudentSubscription(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}
