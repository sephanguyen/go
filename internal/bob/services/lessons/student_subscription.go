package services

import (
	"context"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type StudentSubscriptionService struct {
	bpb.UnimplementedStudentSubscriptionServiceServer

	RetrieveStudentSubscriptionServiceV2   func(context.Context, *lpb.RetrieveStudentSubscriptionRequest) (*lpb.RetrieveStudentSubscriptionResponse, error)
	GetStudentCourseSubscriptionsServiceV2 func(context.Context, *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error)
}

func (s *StudentSubscriptionService) RetrieveStudentSubscription(ctx context.Context, req *bpb.RetrieveStudentSubscriptionRequest) (*bpb.RetrieveStudentSubscriptionResponse, error) {
	lReq := &lpb.RetrieveStudentSubscriptionRequest{
		Paging:     req.GetPaging(),
		Keyword:    req.GetKeyword(),
		LessonDate: req.GetLessonDate(),
	}
	filter := req.GetFilter()
	if filter != nil {
		lReq.Filter = &lpb.RetrieveStudentSubscriptionFilter{
			Grade:      req.Filter.GetGrade(),
			CourseId:   req.Filter.GetCourseId(),
			ClassId:    req.Filter.GetClassId(),
			LocationId: req.Filter.GetLocationId(),
			GradesV2:   req.Filter.GetGradesV2(),
		}
	}
	lRes, err := s.RetrieveStudentSubscriptionServiceV2(ctx, lReq)
	if err != nil {
		return nil, err
	}

	bItems := make([]*bpb.RetrieveStudentSubscriptionResponse_StudentSubscription, 0, len(lRes.GetItems()))

	for _, v := range lRes.GetItems() {
		item := &bpb.RetrieveStudentSubscriptionResponse_StudentSubscription{
			Id:          v.GetId(),
			StudentId:   v.GetStudentId(),
			CourseId:    v.GetCourseId(),
			Grade:       v.GetGrade(),
			LocationIds: v.GetLocationIds(),
			ClassId:     v.GetClassId(),
			StartDate:   v.GetStartDate(),
			EndDate:     v.GetEndDate(),
			GradeV2:     v.GetGradeV2(),
		}
		bItems = append(bItems, item)
	}

	return &bpb.RetrieveStudentSubscriptionResponse{
		Items: bItems,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lRes.NextPage.GetOffsetString(),
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lRes.PreviousPage.GetOffsetString(),
			},
		},
		TotalItems: lRes.GetTotalItems(),
	}, nil
}

func (s *StudentSubscriptionService) GetStudentCourseSubscriptions(ctx context.Context, req *bpb.GetStudentCourseSubscriptionsRequest) (*bpb.GetStudentCourseSubscriptionsResponse, error) {
	lStudentSubscription := make([]*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription, 0, len(req.GetSubscriptions()))

	for _, v := range req.GetSubscriptions() {
		item := &lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
			StudentId: v.GetStudentId(),
			CourseId:  v.GetCourseId(),
		}
		lStudentSubscription = append(lStudentSubscription, item)
	}

	lReq := &lpb.GetStudentCourseSubscriptionsRequest{
		LocationId:    req.GetLocationId(),
		Subscriptions: lStudentSubscription,
	}

	lRes, err := s.GetStudentCourseSubscriptionsServiceV2(ctx, lReq)
	if err != nil {
		return nil, err
	}
	bStudentSubscription := make([]*bpb.GetStudentCourseSubscriptionsResponse_StudentSubscription, 0, len(lRes.GetItems()))
	for _, v := range lRes.GetItems() {
		item := &bpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
			Id:          v.GetId(),
			StudentId:   v.GetStudentId(),
			CourseId:    v.GetCourseId(),
			ClassId:     v.GetClassId(),
			Grade:       v.GetGrade(),
			LocationIds: v.GetLocationIds(),
			StartDate:   v.GetStartDate(),
			EndDate:     v.GetEndDate(),
			GradeV2:     v.GetGradeV2(),
		}
		bStudentSubscription = append(bStudentSubscription, item)
	}

	return &bpb.GetStudentCourseSubscriptionsResponse{
		Items: bStudentSubscription,
	}, nil
}
