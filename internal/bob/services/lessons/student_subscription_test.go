package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRetrieveStudentSubscription(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockStudentSubscriptionServiceV2 := MockStudentSubscriptionServiceV2{}

	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "5",
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

	s := &StudentSubscriptionService{
		RetrieveStudentSubscriptionServiceV2: mockStudentSubscriptionServiceV2.RetrieveStudentSubscriptionServiceV2,
	}

	courses := pgtype.TextArray{}
	_ = courses.Set([]string{"course-1", "course-2"})

	classIds := []string{"class-id-0", "class-id-1"}
	locationIds := []string{"location-id-1"}
	studentIds := []string{"student-1"}
	subIds := []string{"sub-1", "sub-2"}
	lessonDate := time.Date(2022, 6, 20, 4, 0, 0, 0, time.UTC)
	startDate := time.Date(2022, 6, 01, 4, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)

	t.Run("School Admin get list student subs successfully with filter", func(t *testing.T) {
		expectedReq := &bpb.RetrieveStudentSubscriptionRequest{
			Paging:  &cpb.Paging{Limit: 2},
			Keyword: "Student Name",
			Filter: &bpb.RetrieveStudentSubscriptionFilter{
				CourseId:   []string{"course-1", "course-2"},
				Grade:      []string{"5", "6"},
				ClassId:    classIds,
				LocationId: locationIds,
			},
			LessonDate: timestamppb.New(lessonDate),
		}
		expectedRes := &bpb.RetrieveStudentSubscriptionResponse{
			Items: []*bpb.RetrieveStudentSubscriptionResponse_StudentSubscription{
				{
					Id:          subIds[0],
					StudentId:   studentIds[0],
					CourseId:    "course-1",
					Grade:       "5",
					LocationIds: []string{"location-id-1"},
					ClassId:     "class-id-0",
					StartDate:   timestamppb.New(startDate),
					EndDate:     timestamppb.New(endDate),
				},
				{
					Id:        subIds[1],
					StudentId: studentIds[0],
					CourseId:  "course-2",
					Grade:     "5",
					ClassId:   "class-id-1",
					StartDate: timestamppb.New(startDate),
					EndDate:   timestamppb.New(endDate),
				},
			},
			NextPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "sub-2",
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: 2,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "",
				},
			},
			TotalItems: 99,
		}

		lItems := make([]*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription, 0, len(expectedRes.GetItems()))

		for _, v := range expectedRes.GetItems() {
			item := &lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{
				Id:          v.GetId(),
				StudentId:   v.GetStudentId(),
				CourseId:    v.GetCourseId(),
				Grade:       v.GetGrade(),
				LocationIds: v.GetLocationIds(),
				ClassId:     v.GetClassId(),
				StartDate:   v.GetStartDate(),
				EndDate:     v.GetEndDate(),
			}
			lItems = append(lItems, item)
		}

		mockStudentSubscriptionServiceV2.On("RetrieveStudentSubscriptionServiceV2", mock.Anything, mock.MatchedBy(func(lReq *lpb.RetrieveStudentSubscriptionRequest) bool {
			assert.Equal(t, expectedReq.Paging, lReq.Paging)
			assert.Equal(t, expectedReq.Keyword, lReq.Keyword)
			assert.Equal(t, expectedReq.Filter.ClassId, lReq.Filter.ClassId)
			assert.Equal(t, expectedReq.Filter.CourseId, lReq.Filter.CourseId)
			assert.Equal(t, expectedReq.Filter.Grade, lReq.Filter.Grade)
			assert.Equal(t, expectedReq.Filter.LocationId, lReq.Filter.LocationId)

			return true
		})).Once().Return(&lpb.RetrieveStudentSubscriptionResponse{
			Items:        lItems,
			NextPage:     expectedRes.NextPage,
			PreviousPage: expectedRes.PreviousPage,
			TotalItems:   expectedRes.TotalItems,
		}, nil)

		res, err := s.RetrieveStudentSubscription(ctx, expectedReq)

		assert.NoError(t, err)
		assert.Equal(t, expectedRes, res)
	})

	t.Run("fail by call new service has error", func(t *testing.T) {
		mockStudentSubscriptionServiceV2 := MockStudentSubscriptionServiceV2{}
		s := &StudentSubscriptionService{
			RetrieveStudentSubscriptionServiceV2: mockStudentSubscriptionServiceV2.RetrieveStudentSubscriptionServiceV2,
		}
		req := &lpb.RetrieveStudentSubscriptionRequest{}
		expectError := status.Error(codes.Internal, "missing paging info")

		mockStudentSubscriptionServiceV2.On("RetrieveStudentSubscriptionServiceV2", mock.Anything, mock.Anything).Once().Return(nil, expectError)

		_, err := s.RetrieveStudentSubscriptionServiceV2(ctx, req)
		assert.Equal(t, expectError, err)
	})
}

func TestGetStudentCourseSubscriptions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockStudentSubscriptionServiceV2 := MockStudentSubscriptionServiceV2{}

	startDate := time.Date(2022, 6, 27, 4, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)

	s := &StudentSubscriptionService{
		GetStudentCourseSubscriptionsServiceV2: mockStudentSubscriptionServiceV2.GetStudentCourseSubscriptionsServiceV2,
	}

	courses := pgtype.TextArray{}
	_ = courses.Set([]string{"course-1", "course-2"})
	const studentId1 = "student-id-1"

	t.Run("School Admin get list student course successfully with filter", func(t *testing.T) {
		expectedReq := &bpb.GetStudentCourseSubscriptionsRequest{
			Subscriptions: []*bpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
				{
					StudentId: studentId1,
					CourseId:  "course-id-1",
				},
				{
					StudentId: studentId1,
					CourseId:  "course-id-2",
				},
				{
					StudentId: "student-id-3",
					CourseId:  "course-id-5",
				},
			},
			LocationId: "location-id-1",
		}
		expectedRes := &bpb.GetStudentCourseSubscriptionsResponse{
			Items: []*bpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
				{
					Id:          "subscription-id-1",
					StudentId:   studentId1,
					CourseId:    "course-id-1",
					LocationIds: []string{"location-id-1", "location-id-3"},
					StartDate:   timestamppb.New(startDate),
					EndDate:     timestamppb.New(endDate),
				},
				{
					Id:          "subscription-id-2",
					StudentId:   studentId1,
					CourseId:    "course-id-2",
					LocationIds: []string{"location-id-1", "location-id-2", "location-id-5"},
					StartDate:   timestamppb.New(startDate),
					EndDate:     timestamppb.New(endDate),
				},
				{
					Id:        "subscription-id-3",
					StudentId: "student-id-3",
					CourseId:  "course-id-5",
					StartDate: timestamppb.New(startDate),
					EndDate:   timestamppb.New(endDate),
				},
			},
		}

		lItems := make([]*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription, 0, len(expectedRes.GetItems()))

		for _, v := range expectedRes.GetItems() {
			item := &lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
				Id:          v.GetId(),
				StudentId:   v.GetStudentId(),
				CourseId:    v.GetCourseId(),
				Grade:       v.GetGrade(),
				LocationIds: v.GetLocationIds(),
				StartDate:   v.GetStartDate(),
				EndDate:     v.GetEndDate(),
			}
			lItems = append(lItems, item)
		}
		lRes := &lpb.GetStudentCourseSubscriptionsResponse{
			Items: lItems,
		}

		mockStudentSubscriptionServiceV2.On("GetStudentCourseSubscriptionsServiceV2", mock.Anything, mock.MatchedBy(func(lReq *lpb.GetStudentCourseSubscriptionsRequest) bool {
			assert.Equal(t, expectedReq.LocationId, lReq.LocationId)
			for _, v := range expectedReq.Subscriptions {
				assert.Equal(t, true, checkEqualStudentSubscription(v, lReq.Subscriptions))
			}

			return true
		})).Once().Return(lRes, nil)

		res, err := s.GetStudentCourseSubscriptions(ctx, expectedReq)

		assert.NoError(t, err)
		assert.Equal(t, expectedRes, res)
	})

	t.Run("fail by call new service has error", func(t *testing.T) {
		mockStudentSubscriptionServiceV2 := MockStudentSubscriptionServiceV2{}
		s := &StudentSubscriptionService{
			GetStudentCourseSubscriptionsServiceV2: mockStudentSubscriptionServiceV2.GetStudentCourseSubscriptionsServiceV2,
		}
		req := &lpb.GetStudentCourseSubscriptionsRequest{}
		expectError := status.Error(codes.Internal, "studentSubscriptionRepo.GetStudentCourseSubscriptions:")

		mockStudentSubscriptionServiceV2.On("GetStudentCourseSubscriptionsServiceV2", mock.Anything, mock.Anything).Once().Return(nil, expectError)

		_, err := s.GetStudentCourseSubscriptionsServiceV2(ctx, req)
		assert.Equal(t, expectError, err)
	})
}

func checkEqualStudentSubscription(v *bpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription, lc []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription) bool {
	for _, b := range lc {
		if v.CourseId == b.CourseId && v.StudentId == b.StudentId {
			return true
		}
	}
	return false
}

type MockStudentSubscriptionServiceV2 struct {
	mock.Mock
}

func (r *MockStudentSubscriptionServiceV2) RetrieveStudentSubscriptionServiceV2(arg1 context.Context, arg2 *lpb.RetrieveStudentSubscriptionRequest) (*lpb.RetrieveStudentSubscriptionResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.RetrieveStudentSubscriptionResponse), args.Error(1)
}

func (r *MockStudentSubscriptionServiceV2) GetStudentCourseSubscriptionsServiceV2(arg1 context.Context, arg2 *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lpb.GetStudentCourseSubscriptionsResponse), args.Error(1)
}
