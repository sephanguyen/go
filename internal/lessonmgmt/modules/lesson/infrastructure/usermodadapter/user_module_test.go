package usermodadapter

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGRPCServiceMock struct {
	getStudentsManyReferenceByNameOrEmail func(context.Context, *lpb.GetStudentsManyReferenceByNameOrEmailRequest) (*lpb.GetStudentsManyReferenceByNameOrEmailResponse, error)
	getTeachers                           func(context.Context, *lpb.GetTeachersRequest) (*lpb.GetTeachersResponse, error)
	getUserGroup                          func(ctx context.Context, userID *lpb.GetUserGroupRequest) (*lpb.GetUserGroupResponse, error)
	getTeachersSameGrantedLocation        func(context.Context, *lpb.GetTeachersSameGrantedLocationRequest) (*lpb.GetTeachersSameGrantedLocationResponse, error)
}

func (u *UserGRPCServiceMock) GetStudentsManyReferenceByNameOrEmail(ctx context.Context, request *lpb.GetStudentsManyReferenceByNameOrEmailRequest) (*lpb.GetStudentsManyReferenceByNameOrEmailResponse, error) {
	return u.getStudentsManyReferenceByNameOrEmail(ctx, request)
}

func (u *UserGRPCServiceMock) GetTeachers(ctx context.Context, req *lpb.GetTeachersRequest) (*lpb.GetTeachersResponse, error) {
	return u.getTeachers(ctx, req)
}

func (u *UserGRPCServiceMock) GetUserGroup(ctx context.Context, userID *lpb.GetUserGroupRequest) (*lpb.GetUserGroupResponse, error) {
	return u.getUserGroup(ctx, userID)
}

func (u *UserGRPCServiceMock) GetTeachersSameGrantedLocation(ctx context.Context, req *lpb.GetTeachersSameGrantedLocationRequest) (*lpb.GetTeachersSameGrantedLocationResponse, error) {
	return u.getTeachersSameGrantedLocation(ctx, req)
}

type StudentSubscriptionServiceMock struct {
	getStudentCourseSubscriptions    func(context.Context, *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error)
	retrieveStudentSubscription      func(context.Context, *lpb.RetrieveStudentSubscriptionRequest) (*lpb.RetrieveStudentSubscriptionResponse, error)
	retrieveStudentPendingReallocate func(context.Context, *lpb.RetrieveStudentPendingReallocateRequest) (*lpb.RetrieveStudentPendingReallocateResponse, error)
	getStudentCoursesAndClasses      func(ctx context.Context, req *lpb.GetStudentCoursesAndClassesRequest) (*lpb.GetStudentCoursesAndClassesResponse, error)
}

func (s *StudentSubscriptionServiceMock) GetStudentCourseSubscriptions(ctx context.Context, req *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
	return s.getStudentCourseSubscriptions(ctx, req)
}

func (s *StudentSubscriptionServiceMock) RetrieveStudentSubscription(ctx context.Context, req *lpb.RetrieveStudentSubscriptionRequest) (*lpb.RetrieveStudentSubscriptionResponse, error) {
	return s.retrieveStudentSubscription(ctx, req)
}

func (s *StudentSubscriptionServiceMock) RetrieveStudentPendingReallocate(ctx context.Context, req *lpb.RetrieveStudentPendingReallocateRequest) (*lpb.RetrieveStudentPendingReallocateResponse, error) {
	return s.retrieveStudentPendingReallocate(ctx, req)
}

func (s *StudentSubscriptionServiceMock) GetStudentCoursesAndClasses(ctx context.Context, req *lpb.GetStudentCoursesAndClassesRequest) (*lpb.GetStudentCoursesAndClassesResponse, error) {
	return s.getStudentCoursesAndClasses(ctx, req)
}

func TestUserModule_CheckTeacherIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tcs := []struct {
		name     string
		teachIDs []string
		*UserGRPCServiceMock
		hasError bool
	}{
		{
			name:     "all teacher ids exist",
			teachIDs: []string{"user-id-1", "user-id-2", "user-id-3"},
			UserGRPCServiceMock: &UserGRPCServiceMock{
				getTeachers: func(ctx context.Context, request *lpb.GetTeachersRequest) (*lpb.GetTeachersResponse, error) {
					assert.Equal(t, []string{"user-id-1", "user-id-2", "user-id-3"}, request.TeacherIds)
					return &lpb.GetTeachersResponse{
						Teachers: []*lpb.GetTeachersResponse_TeacherInfo{
							{
								Id: "user-id-1",
							},
							{
								Id: "user-id-2",
							},
							{
								Id: "user-id-3",
							},
						},
					}, nil
				},
			},
		},
		{
			name:     "duplicated input id",
			teachIDs: []string{"user-id-1", "user-id-2", "user-id-3", "user-id-1"},
			UserGRPCServiceMock: &UserGRPCServiceMock{
				getTeachers: func(ctx context.Context, request *lpb.GetTeachersRequest) (*lpb.GetTeachersResponse, error) {
					assert.Equal(t, []string{"user-id-1", "user-id-2", "user-id-3", "user-id-1"}, request.TeacherIds)
					return &lpb.GetTeachersResponse{
						Teachers: []*lpb.GetTeachersResponse_TeacherInfo{
							{
								Id: "user-id-1",
							},
							{
								Id: "user-id-2",
							},
							{
								Id: "user-id-3",
							},
						},
					}, nil
				},
			},
		},
		{
			name:     "could not found some teacher ids",
			teachIDs: []string{"user-id-1", "user-id-2", "user-id-3"},
			UserGRPCServiceMock: &UserGRPCServiceMock{
				getTeachers: func(ctx context.Context, request *lpb.GetTeachersRequest) (*lpb.GetTeachersResponse, error) {
					assert.Equal(t, []string{"user-id-1", "user-id-2", "user-id-3"}, request.TeacherIds)
					return &lpb.GetTeachersResponse{
						Teachers: []*lpb.GetTeachersResponse_TeacherInfo{
							{
								Id: "user-id-1",
							},
							{
								Id: "user-id-3",
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			u := &UserModuleAdapter{
				Module: &user.Module{
					UserGRPCService: tc.UserGRPCServiceMock,
				},
			}
			err := u.CheckTeacherIDs(ctx, tc.teachIDs)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserModule_CheckStudentCourseSubscriptions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// startDate and endDate aligned with lessonDate
	startDate := timestamppb.New(time.Date(2022, 8, 12, 9, 0, 0, 0, time.UTC))
	endDate := timestamppb.New(time.Date(2022, 9, 12, 9, 0, 0, 0, time.UTC))
	tcs := []struct {
		name                 string
		studentWithCourseIDs []string
		locationID           string
		*StudentSubscriptionServiceMock
		hasError bool
	}{
		{
			name: "there are all student course subscriptions",
			studentWithCourseIDs: []string{
				"user-id-1",
				"course-id-1",
				"user-id-1",
				"course-id-2",
				"user-id-3",
				"course-id-5",
			},
			locationID: "location-id-1",
			StudentSubscriptionServiceMock: &StudentSubscriptionServiceMock{
				getStudentCourseSubscriptions: func(ctx context.Context, request *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
					assert.Equal(t, &lpb.GetStudentCourseSubscriptionsRequest{
						Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
							},
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
							},
							{
								StudentId: "user-id-3",
								CourseId:  "course-id-5",
							},
						},
						//LocationId: "location-id-1",
					}, request)
					return &lpb.GetStudentCourseSubscriptionsResponse{
						Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
							{
								Id:        "subscription-id-1",
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
								StartDate: startDate,
								EndDate:   endDate,
							},
							{
								Id:        "subscription-id-2",
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
								StartDate: startDate,
								EndDate:   endDate,
							},
							{
								Id:        "subscription-id-3",
								StudentId: "user-id-3",
								CourseId:  "course-id-5",
								StartDate: startDate,
								EndDate:   endDate,
							},
						},
					}, nil
				},
			},
		},
		{
			name: "duplicated student course subscriptions input",
			studentWithCourseIDs: []string{
				"user-id-1",
				"course-id-1",
				"user-id-1",
				"course-id-2",
				"user-id-3",
				"course-id-5",
				"user-id-1",
				"course-id-2",
			},
			locationID: "location-id-1",
			StudentSubscriptionServiceMock: &StudentSubscriptionServiceMock{
				getStudentCourseSubscriptions: func(ctx context.Context, request *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
					assert.Equal(t, &lpb.GetStudentCourseSubscriptionsRequest{
						Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
							},
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
							},
							{
								StudentId: "user-id-3",
								CourseId:  "course-id-5",
							},
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
							},
						},
						//LocationId: "location-id-1",
					}, request)
					return &lpb.GetStudentCourseSubscriptionsResponse{
						Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
							{
								Id:        "subscription-id-1",
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
								StartDate: startDate,
								EndDate:   endDate,
							},
							{
								Id:        "subscription-id-2",
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
								StartDate: startDate,
								EndDate:   endDate,
							},
							{
								Id:        "subscription-id-3",
								StudentId: "user-id-3",
								CourseId:  "course-id-5",
								StartDate: startDate,
								EndDate:   endDate,
							},
						},
					}, nil
				},
			},
		},
		{
			name: "could not found some student course subscriptions",
			studentWithCourseIDs: []string{
				"user-id-1",
				"course-id-1",
				"user-id-1",
				"course-id-2",
				"user-id-3",
				"course-id-5",
			},
			locationID: "location-id-1",
			StudentSubscriptionServiceMock: &StudentSubscriptionServiceMock{
				getStudentCourseSubscriptions: func(ctx context.Context, request *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
					assert.Equal(t, &lpb.GetStudentCourseSubscriptionsRequest{
						Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
							},
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
							},
							{
								StudentId: "user-id-3",
								CourseId:  "course-id-5",
							},
						},
						//LocationId: "location-id-1",
					}, request)
					return &lpb.GetStudentCourseSubscriptionsResponse{
						Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
							{
								Id:        "subscription-id-1",
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
							},
							{
								Id:        "subscription-id-3",
								StudentId: "user-id-3",
								CourseId:  "course-id-5",
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
		{
			name: "not enough arguments",
			studentWithCourseIDs: []string{
				"user-id-1",
				"course-id-1",
				"user-id-1",
				"course-id-2",
				"user-id-3",
			},
			locationID: "location-id-1",
			StudentSubscriptionServiceMock: &StudentSubscriptionServiceMock{
				getStudentCourseSubscriptions: func(ctx context.Context, request *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
					assert.Equal(t, &lpb.GetStudentCourseSubscriptionsRequest{
						Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
							},
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
							},
							{
								StudentId: "user-id-3",
								CourseId:  "",
							},
						},
						//LocationId: "location-id-1",
					}, request)
					return &lpb.GetStudentCourseSubscriptionsResponse{
						Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
							{
								Id:        "subscription-id-1",
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
							},
							{
								Id:        "subscription-id-2",
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
		{
			name: "at least one student subscription duration not aligned with lesson date",
			studentWithCourseIDs: []string{
				"user-id-1",
				"course-id-1",
				"user-id-1",
				"course-id-2",
			},
			locationID: "location-id-1",
			StudentSubscriptionServiceMock: &StudentSubscriptionServiceMock{
				getStudentCourseSubscriptions: func(ctx context.Context, request *lpb.GetStudentCourseSubscriptionsRequest) (*lpb.GetStudentCourseSubscriptionsResponse, error) {
					assert.Equal(t, &lpb.GetStudentCourseSubscriptionsRequest{
						Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
							},
							{
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
							},
						},
						//LocationId: "location-id-1",
					}, request)
					return &lpb.GetStudentCourseSubscriptionsResponse{
						Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
							{
								Id:        "subscription-id-1",
								StudentId: "user-id-1",
								CourseId:  "course-id-1",
								StartDate: timestamppb.New(time.Date(2022, 8, 21, 9, 0, 0, 0, time.UTC)),
								EndDate:   timestamppb.New(time.Date(2022, 9, 12, 9, 0, 0, 0, time.UTC)),
							},
							{
								Id:        "subscription-id-2",
								StudentId: "user-id-1",
								CourseId:  "course-id-2",
								StartDate: timestamppb.New(time.Date(2022, 8, 12, 9, 0, 0, 0, time.UTC)),
								EndDate:   timestamppb.New(time.Date(2022, 8, 19, 9, 0, 0, 0, time.UTC)),
							},
						},
					}, nil
				},
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			u := &UserModuleAdapter{
				Module: &user.Module{
					StudentSubscriptionGRPCLessonmgmtService: tc.StudentSubscriptionServiceMock,
				},
			}
			err := u.CheckStudentCourseSubscriptions(ctx, time.Date(2022, 8, 20, 9, 0, 0, 0, time.UTC), tc.studentWithCourseIDs...)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
