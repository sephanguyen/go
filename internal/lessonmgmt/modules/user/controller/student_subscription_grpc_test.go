package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"
	mock_master_repositories "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStudentSubscriptionGRPCService_GetStudentCourseSubscriptions(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	studentSubscriptionRepo := new(mock_repositories.MockStudentSubscriptionRepo)
	classRepo := new(mock_master_repositories.MockClassRepo)

	tcs := []struct {
		name     string
		req      *lpb.GetStudentCourseSubscriptionsRequest
		res      *lpb.GetStudentCourseSubscriptionsResponse
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "get successfully",
			req: &lpb.GetStudentCourseSubscriptionsRequest{
				Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
					{
						StudentId: "student-id-1",
						CourseId:  "course-id-1",
					},
					{
						StudentId: "student-id-1",
						CourseId:  "course-id-2",
					},
					{
						StudentId: "student-id-3",
						CourseId:  "course-id-5",
					},
				},
				LocationId: "location-id-1",
			},
			res: &lpb.GetStudentCourseSubscriptionsResponse{
				Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
					{
						Id:          "subscription-id-1",
						StudentId:   "student-id-1",
						CourseId:    "course-id-1",
						LocationIds: []string{"location-id-1", "location-id-3"},
						StartDate:   timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:     timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
					{
						Id:          "subscription-id-2",
						StudentId:   "student-id-1",
						CourseId:    "course-id-2",
						LocationIds: []string{"location-id-1", "location-id-2", "location-id-5"},
						StartDate:   timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:     timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
					{
						Id:        "subscription-id-3",
						StudentId: "student-id-3",
						CourseId:  "course-id-5",
						StartDate: timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:   timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				studentSubscription := domain.StudentSubscriptions{
					{
						SubscriptionID: "subscription-id-1",
						StudentID:      "student-id-1",
						CourseID:       "course-id-1",
						LocationIDs:    []string{"location-id-1", "location-id-3"},
						StartAt:        time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:          time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
					{
						SubscriptionID: "subscription-id-2",
						StudentID:      "student-id-1",
						CourseID:       "course-id-2",
						LocationIDs:    []string{"location-id-1", "location-id-2", "location-id-5"},
						StartAt:        time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:          time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
					{
						SubscriptionID: "subscription-id-3",
						StudentID:      "student-id-3",
						CourseID:       "course-id-5",
						StartAt:        time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:          time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On(
						"GetStudentCourseSubscriptions",
						ctx,
						db,
						[]string{"location-id-1"},
						[]string{
							"student-id-1",
							"course-id-1",
							"student-id-1",
							"course-id-2",
							"student-id-3",
							"course-id-5",
						},
					).
					Return(studentSubscription, nil).Once()
				studentCoursesReq := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				studentCoursesRes := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				for _, sub := range studentSubscription {
					scReq := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesReq = append(studentCoursesReq, scReq)
					scRes := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesRes = append(studentCoursesRes, scRes)
				}

				classRepo.On("FindByCourseIDsAndStudentIDs", mock.Anything, mock.Anything, studentCoursesReq).Once().Return(studentCoursesRes, nil)

			},
		},
		{
			name: "could not found some subscription",
			req: &lpb.GetStudentCourseSubscriptionsRequest{
				Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
					{
						StudentId: "student-id-1",
						CourseId:  "course-id-1",
					},
					{
						StudentId: "student-id-1",
						CourseId:  "course-id-2",
					},
					{
						StudentId: "student-id-3",
						CourseId:  "course-id-5",
					},
				},
			},
			res: &lpb.GetStudentCourseSubscriptionsResponse{
				Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
					{
						Id:          "subscription-id-1",
						StudentId:   "student-id-1",
						CourseId:    "course-id-1",
						LocationIds: []string{"location-id-1", "location-id-3"},
						StartDate:   timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:     timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
					{
						Id:          "subscription-id-3",
						StudentId:   "student-id-3",
						CourseId:    "course-id-5",
						LocationIds: []string{"location-id-1", "location-id-2", "location-id-5"},
						StartDate:   timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:     timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				studentSubscription := domain.StudentSubscriptions{
					{
						SubscriptionID: "subscription-id-1",
						StudentID:      "student-id-1",
						CourseID:       "course-id-1",
						LocationIDs:    []string{"location-id-1", "location-id-3"},
						StartAt:        time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:          time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
					{
						SubscriptionID: "subscription-id-3",
						StudentID:      "student-id-3",
						CourseID:       "course-id-5",
						LocationIDs:    []string{"location-id-1", "location-id-2", "location-id-5"},
						StartAt:        time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:          time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On(
						"GetStudentCourseSubscriptions",
						ctx,
						db,
						mock.Anything,
						[]string{
							"student-id-1",
							"course-id-1",
							"student-id-1",
							"course-id-2",
							"student-id-3",
							"course-id-5",
						},
					).
					Return(studentSubscription, nil).Once()
				studentCoursesReq := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				studentCoursesRes := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				for _, sub := range studentSubscription {
					scReq := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesReq = append(studentCoursesReq, scReq)
					scRes := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesRes = append(studentCoursesRes, scRes)
				}
				classRepo.On("FindByCourseIDsAndStudentIDs", mock.Anything, mock.Anything, studentCoursesReq).Once().Return(studentCoursesRes, nil)

			},
		},
		{
			name: "return invalid subscription",
			req: &lpb.GetStudentCourseSubscriptionsRequest{
				Subscriptions: []*lpb.GetStudentCourseSubscriptionsRequest_StudentCourseSubscription{
					{
						StudentId: "student-id-1",
						CourseId:  "course-id-1",
					},
					{
						StudentId: "student-id-1",
						CourseId:  "course-id-2",
					},
					{
						StudentId: "student-id-3",
						CourseId:  "course-id-5",
					},
				},
				LocationId: "location-id-1",
			},
			res: &lpb.GetStudentCourseSubscriptionsResponse{
				Items: []*lpb.GetStudentCourseSubscriptionsResponse_StudentSubscription{
					{
						Id:        "subscription-id-1",
						StudentId: "student-id-1",
						CourseId:  "course-id-1",
						StartDate: timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:   timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
					{
						Id:        "subscription-id-2",
						StudentId: "student-id-1",
						CourseId:  "course-id-2",
						StartDate: timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:   timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
					{
						Id:        "subscription-id-3",
						StudentId: "student-id-3",
						CourseId:  "course-id-5",
						StartDate: timestamppb.New(time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC)),
						EndDate:   timestamppb.New(time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC)),
					},
				},
			},
			setup: func(ctx context.Context) {
				studentSubscription := domain.StudentSubscriptions{
					{
						SubscriptionID: "subscription-id-1",
						StudentID:      "student-id-1",
						CourseID:       "course-id-1",
						StartAt:        time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:          time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
					{
						StudentID: "student-id-1",
						CourseID:  "course-id-2",
						StartAt:   time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:     time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
					{
						SubscriptionID: "subscription-id-3",
						StudentID:      "student-id-3",
						CourseID:       "course-id-5",
						StartAt:        time.Date(2022, 8, 15, 9, 0, 0, 0, time.UTC),
						EndAt:          time.Date(2022, 9, 15, 9, 0, 0, 0, time.UTC),
					},
				}
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On(
						"GetStudentCourseSubscriptions",
						ctx,
						db,
						[]string{"location-id-1"},
						[]string{
							"student-id-1",
							"course-id-1",
							"student-id-1",
							"course-id-2",
							"student-id-3",
							"course-id-5",
						},
					).
					Return(studentSubscription, nil).Once()
				studentCoursesReq := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				studentCoursesRes := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				for _, sub := range studentSubscription {
					scReq := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesReq = append(studentCoursesReq, scReq)
					scRes := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesRes = append(studentCoursesRes, scRes)
				}
				classRepo.On("FindByCourseIDsAndStudentIDs", mock.Anything, mock.Anything, studentCoursesReq).Once().Return(studentCoursesRes, nil)

			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			src := NewStudentSubscriptionGRPCService(wrapperConnection, studentSubscriptionRepo, nil, nil, classRepo, "", nil)
			actual, err := src.GetStudentCourseSubscriptions(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.res, actual)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				studentSubscriptionRepo,
				mockUnleashClient,
			)
		})
	}
}

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestStudentSubscriptionGRPCService_RetrieveStudentSubscription(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentSubscriptionRepo := new(mock_repositories.MockStudentSubscriptionRepo)
	studentSubscriptionAccessPathRepo := new(mock_repositories.MockStudentSubscriptionAccessPathRepo)
	classMemberRepo := new(mock_master_repositories.MockClassMemberRepo)
	classRepo := new(mock_master_repositories.MockClassRepo)
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "5",
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

	s := NewStudentSubscriptionGRPCService(wrapperConnection, studentSubscriptionRepo, studentSubscriptionAccessPathRepo, classMemberRepo, classRepo, "", mockUnleashClient)

	courses := []string{"course-1", "course-2"}

	classIds := []string{"class-id-0", "class-id-1"}
	locationIds := []string{"location-id-1"}
	studentIds := []string{"student-1"}
	studentIdCourseIds := []string{"student-1", "course-1"}
	subIds := []string{"sub-1", "sub-2"}
	lessonDate := time.Date(2022, 6, 27, 9, 0, 0, 0, time.UTC)
	startDate := time.Date(2022, 6, 27, 4, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, 6, 27, 10, 0, 0, 0, time.UTC)

	testCases := []TestCase{
		{
			name: "School Admin get list student subs successfully with filter",
			req: &lpb.RetrieveStudentSubscriptionRequest{
				Paging:  &cpb.Paging{Limit: 2},
				Keyword: "Student Name",
				Filter: &lpb.RetrieveStudentSubscriptionFilter{
					CourseId:   courses,
					Grade:      []string{"5", "6"},
					ClassId:    classIds,
					LocationId: locationIds,
				},
				LessonDate: timestamppb.New(lessonDate),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveStudentSubscriptionResponse{
				Items: []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{
					{
						Id:          subIds[0],
						StudentId:   studentIds[0],
						CourseId:    courses[0],
						Grade:       "5",
						LocationIds: locationIds,
						ClassId:     classIds[0],
						StartDate:   timestamppb.New(startDate),
						EndDate:     timestamppb.New(endDate),
					},
					{
						Id:        subIds[1],
						StudentId: studentIds[0],
						CourseId:  courses[1],
						Grade:     "5",
						ClassId:   classIds[1],
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
			},
			setup: func(ctx context.Context) {
				studentSubscription := []*domain.StudentSubscription{
					{
						SubscriptionID: "sub-1",
						CourseID:       courses[0],
						StudentID:      studentIds[0],
						Grade:          "5",
						StartAt:        startDate,
						EndAt:          endDate,
					},
					{
						SubscriptionID: "sub-2",
						CourseID:       courses[1],
						StudentID:      studentIds[0],
						Grade:          "5",
						StartAt:        startDate,
						EndAt:          endDate,
					},
				}

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("FindStudentIDWithCourseIDsByClassIDs", mock.Anything, mock.Anything, classIds).Once().Return(studentIdCourseIds, nil)
				studentSubscriptionAccessPathRepo.On("FindStudentSubscriptionIDsByLocationIDs", mock.Anything, mock.Anything, locationIds).Once().Return(subIds, nil)
				studentSubscriptionRepo.On("RetrieveStudentSubscription", mock.Anything, mock.Anything, &payloads.ListStudentSubScriptionsArgs{
					Limit:                  2,
					SchoolID:               "5",
					Grades:                 []int32{5, 6},
					KeyWord:                "Student Name",
					CourseIDs:              courses,
					StudentSubscriptionID:  "",
					ClassIDs:               classIds,
					LocationIDs:            locationIds,
					StudentIDWithCourseIDs: studentIdCourseIds,
					StudentSubscriptionIDs: subIds,
					LessonDate:             lessonDate,
				}).Once().Return(studentSubscription, uint32(99), "pre_id", uint32(2), nil)
				studentSubscriptionAccessPathRepo.On("FindLocationsByStudentSubscriptionIDs", mock.Anything, mock.Anything, []string{"sub-1", "sub-2"}).Once().Return(
					map[string][]string{
						"sub-1": {"location-id-1"},
					}, nil,
				)
				studentCoursesReq := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				studentCoursesRes := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				for _, sub := range studentSubscription {
					scReq := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesReq = append(studentCoursesReq, scReq)
					scRes := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesRes = append(studentCoursesRes, scRes)
				}

				studentCoursesRes[0].ClassID = classIds[0]
				studentCoursesRes[1].ClassID = classIds[1]
				classRepo.On("FindByCourseIDsAndStudentIDs", mock.Anything, mock.Anything, studentCoursesReq).Once().Return(studentCoursesRes, nil)
			},
		},
		{
			name: "School Admin get list student subs successfully without filter",
			req: &lpb.RetrieveStudentSubscriptionRequest{
				Paging:     &cpb.Paging{Limit: 2},
				LessonDate: timestamppb.New(lessonDate),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveStudentSubscriptionResponse{
				Items: []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{
					{
						Id:        "sub-1",
						StudentId: studentIds[0],
						CourseId:  courses[0],
						Grade:     "5",
						ClassId:   classIds[0],
						StartDate: timestamppb.New(startDate),
						EndDate:   timestamppb.New(endDate),
					},
					{
						Id:        "sub-2",
						StudentId: studentIds[0],
						CourseId:  courses[1],
						Grade:     "5",
						ClassId:   classIds[1],
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
			},
			setup: func(ctx context.Context) {
				studentSubscription := []*domain.StudentSubscription{
					{
						SubscriptionID: "sub-1",
						CourseID:       courses[0],
						StudentID:      studentIds[0],
						Grade:          "5",
						StartAt:        startDate,
						EndAt:          endDate,
					},
					{
						SubscriptionID: "sub-2",
						CourseID:       courses[1],
						StudentID:      studentIds[0],
						Grade:          "5",
						StartAt:        startDate,
						EndAt:          endDate,
					},
				}

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.On("RetrieveStudentSubscription", mock.Anything, mock.Anything, &payloads.ListStudentSubScriptionsArgs{
					Limit:      2,
					SchoolID:   "5",
					LessonDate: lessonDate,
				}).Once().Return(studentSubscription, uint32(99), "pre_id", uint32(2), nil)
				studentSubscriptionAccessPathRepo.On("FindLocationsByStudentSubscriptionIDs", mock.Anything, mock.Anything, []string{"sub-1", "sub-2"}).Once().Return(
					map[string][]string{}, nil,
				)
				studentCoursesReq := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				studentCoursesRes := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				for _, sub := range studentSubscription {
					scReq := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesReq = append(studentCoursesReq, scReq)
					scRes := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesRes = append(studentCoursesRes, scRes)
				}

				studentCoursesRes[0].ClassID = classIds[0]
				studentCoursesRes[1].ClassID = classIds[1]
				classRepo.On("FindByCourseIDsAndStudentIDs", mock.Anything, mock.Anything, studentCoursesReq).Once().Return(studentCoursesRes, nil)
			},
		},
		{
			name: "Teacher get list student subs successfully with filter",
			req: &lpb.RetrieveStudentSubscriptionRequest{
				Paging:  &cpb.Paging{Limit: 2},
				Keyword: "Student Name",
				Filter: &lpb.RetrieveStudentSubscriptionFilter{
					CourseId:   courses,
					Grade:      []string{"5", "6"},
					ClassId:    classIds,
					LocationId: locationIds,
				},
				LessonDate: timestamppb.New(lessonDate),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveStudentSubscriptionResponse{
				Items: []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{
					{
						Id:        "sub-1",
						StudentId: studentIds[0],
						CourseId:  courses[0],
						Grade:     "5",
						ClassId:   classIds[0],
						StartDate: timestamppb.New(startDate),
						EndDate:   timestamppb.New(endDate),
					},
					{
						Id:        "sub-2",
						StudentId: studentIds[0],
						CourseId:  courses[1],
						Grade:     "5",
						ClassId:   classIds[1],
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
			},
			setup: func(ctx context.Context) {
				studentSubscription := []*domain.StudentSubscription{
					{
						SubscriptionID: "sub-1",
						CourseID:       courses[0],
						StudentID:      studentIds[0],
						Grade:          "5",
						StartAt:        startDate,
						EndAt:          endDate,
					},
					{
						SubscriptionID: "sub-2",
						CourseID:       courses[1],
						StudentID:      studentIds[0],
						Grade:          "5",
						StartAt:        startDate,
						EndAt:          endDate,
					},
				}

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("FindStudentIDWithCourseIDsByClassIDs", mock.Anything, mock.Anything, classIds).Once().Return(studentIdCourseIds, nil)
				studentSubscriptionAccessPathRepo.On("FindStudentSubscriptionIDsByLocationIDs", mock.Anything, mock.Anything, locationIds).Once().Return(subIds, nil)
				studentSubscriptionRepo.On("RetrieveStudentSubscription", mock.Anything, mock.Anything, &payloads.ListStudentSubScriptionsArgs{
					Limit:                  2,
					SchoolID:               "5",
					Grades:                 []int32{5, 6},
					KeyWord:                "Student Name",
					CourseIDs:              courses,
					ClassIDs:               classIds,
					LocationIDs:            locationIds,
					StudentIDWithCourseIDs: studentIdCourseIds,
					StudentSubscriptionIDs: subIds,
					LessonDate:             lessonDate,
				}).Once().Return(studentSubscription, uint32(99), "pre_id", uint32(2), nil)
				studentSubscriptionAccessPathRepo.On("FindLocationsByStudentSubscriptionIDs", mock.Anything, mock.Anything, []string{"sub-1", "sub-2"}).Once().Return(
					map[string][]string{}, nil,
				)
				studentCoursesReq := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				studentCoursesRes := make([]*class_domain.ClassWithCourseStudent, 0, len(studentSubscription))
				for _, sub := range studentSubscription {
					scReq := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesReq = append(studentCoursesReq, scReq)
					scRes := &class_domain.ClassWithCourseStudent{CourseID: sub.CourseID, StudentID: sub.StudentID}
					studentCoursesRes = append(studentCoursesRes, scRes)
				}

				studentCoursesRes[0].ClassID = classIds[0]
				studentCoursesRes[1].ClassID = classIds[1]
				classRepo.On("FindByCourseIDsAndStudentIDs", mock.Anything, mock.Anything, studentCoursesReq).Once().Return(studentCoursesRes, nil)
			},
		},
		{
			name: "School Admin get list student subs successfully empty",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveStudentSubscriptionRequest{
				Paging:  &cpb.Paging{Limit: 2},
				Keyword: "Student Name",
				Filter: &lpb.RetrieveStudentSubscriptionFilter{
					CourseId:   []string{courses[0], courses[1]},
					Grade:      []string{"1", "6"},
					LocationId: locationIds,
					ClassId:    classIds,
				},
				LessonDate: timestamppb.New(lessonDate),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveStudentSubscriptionResponse{
				Items: []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalItems: 0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("FindStudentIDWithCourseIDsByClassIDs", mock.Anything, mock.Anything, classIds).Once().Return(studentIdCourseIds, nil)
				studentSubscriptionAccessPathRepo.On("FindStudentSubscriptionIDsByLocationIDs", mock.Anything, mock.Anything, locationIds).Once().Return(subIds, nil)
				studentSubscriptionRepo.On("RetrieveStudentSubscription", mock.Anything, mock.Anything, &payloads.ListStudentSubScriptionsArgs{
					Limit:                  2,
					SchoolID:               "5",
					Grades:                 []int32{1, 6},
					KeyWord:                "Student Name",
					CourseIDs:              courses,
					ClassIDs:               classIds,
					LocationIDs:            locationIds,
					StudentIDWithCourseIDs: studentIdCourseIds,
					StudentSubscriptionIDs: subIds,
					LessonDate:             lessonDate,
				}).Once().Return([]*domain.StudentSubscription{}, uint32(0), "", uint32(2), nil)
			},
		},
		{
			name: "School Admin retrieves an empty list of student subs successfully by center which doesn't have a lesson subscription",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveStudentSubscriptionRequest{
				Paging:  &cpb.Paging{Limit: 2},
				Keyword: "Student Name",
				Filter: &lpb.RetrieveStudentSubscriptionFilter{
					CourseId:   []string{courses[0], courses[1]},
					Grade:      []string{"1", "6"},
					LocationId: locationIds,
					ClassId:    classIds,
				},
				LessonDate: timestamppb.New(lessonDate),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveStudentSubscriptionResponse{
				Items: []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalItems: 0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("FindStudentIDWithCourseIDsByClassIDs", mock.Anything, mock.Anything, classIds).Once().Return(studentIdCourseIds, nil)
				studentSubscriptionAccessPathRepo.On("FindStudentSubscriptionIDsByLocationIDs", mock.Anything, mock.Anything, locationIds).Once().Return([]string{}, nil)
			},
		},
		{
			name: "School Admin successfully retrieves an empty list of student subs by class which doesn't have a lesson_member",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.RetrieveStudentSubscriptionRequest{
				Paging:  &cpb.Paging{Limit: 2},
				Keyword: "Student Name",
				Filter: &lpb.RetrieveStudentSubscriptionFilter{
					CourseId:   []string{courses[0], courses[1]},
					Grade:      []string{"1", "6"},
					LocationId: locationIds,
					ClassId:    classIds,
				},
				LessonDate: timestamppb.New(lessonDate),
			},
			expectedErr: nil,
			expectedResp: &lpb.RetrieveStudentSubscriptionResponse{
				Items: []*lpb.RetrieveStudentSubscriptionResponse_StudentSubscription{},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "",
					},
				},
				TotalItems: 0,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("FindStudentIDWithCourseIDsByClassIDs", mock.Anything, mock.Anything, classIds).Once().Return([]string{}, nil)
			},
		},
		{
			name:         "Return fail missing page",
			req:          &lpb.RetrieveStudentSubscriptionRequest{},
			expectedErr:  status.Error(codes.Internal, "missing paging info"),
			expectedResp: &lpb.RetrieveStudentSubscriptionResponse{},
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*lpb.RetrieveStudentSubscriptionRequest)
			resp, err := s.RetrieveStudentSubscription(ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, db, studentSubscriptionRepo, studentSubscriptionAccessPathRepo, classMemberRepo, classRepo, mockUnleashClient)
		})
	}
}

func TestStudentSubscriptionGRPCService_RetrieveStudentPendingReallocate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := &mock_unleash_client.UnleashClientInstance{}
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	studentSubscriptionRepo := new(mock_repositories.MockStudentSubscriptionRepo)

	testCases := []struct {
		name   string
		setup  func(context.Context)
		hasErr bool
	}{
		{
			name: "success",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.On("RetrieveStudentPendingReallocate", mock.Anything, mock.Anything, domain.RetrieveStudentPendingReallocateDto{
					Limit:      5,
					Timezone:   "UTC",
					LessonDate: time.Date(2022, 11, 14, 0, 0, 0, 0, time.UTC),
					StartDate:  time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
					EndDate:    time.Date(2022, 11, 30, 23, 59, 59, 0, time.UTC),
				}).Once().Return([]*domain.ReallocateStudent{
					{
						StudentId:        "student-1",
						OriginalLessonID: "lesson-1",
						CourseID:         "course-1",
						LocationID:       "location-1",
						GradeID:          "grade-1",
						StartAt:          time.Now(),
						EndAt:            time.Now(),
					},
					{
						StudentId:        "student-2",
						OriginalLessonID: "lesson-1",
						CourseID:         "course-1",
						LocationID:       "location-1",
						GradeID:          "grade-1",
						StartAt:          time.Now(),
						EndAt:            time.Now(),
					},
				}, uint32(1), nil)
			},
		},
		{
			name: "something went wrong while query to database",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.On("RetrieveStudentPendingReallocate", mock.Anything, mock.Anything, domain.RetrieveStudentPendingReallocateDto{
					Limit:      5,
					Timezone:   "UTC",
					LessonDate: time.Date(2022, 11, 14, 0, 0, 0, 0, time.UTC),
					StartDate:  time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
					EndDate:    time.Date(2022, 11, 30, 23, 59, 59, 0, time.UTC),
				}).Once().Return(nil, uint32(0), errors.New("something went wrong"))
			},
			hasErr: true,
		},
	}

	service := StudentSubscriptionGRPCService{
		StudentReallocate: queries.StudentReallocateQueryHandler{
			WrapperConnection:       wrapperConnection,
			StudentSubscriptionRepo: studentSubscriptionRepo,
			Env:                     "local",
			UnleashClientIns:        mockUnleashClient,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := service.RetrieveStudentPendingReallocate(ctx, &lpb.RetrieveStudentPendingReallocateRequest{
				Paging: &cpb.Paging{
					Limit: 5,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				LessonDate: timestamppb.New(time.Date(2022, 11, 14, 0, 0, 0, 0, time.UTC)),
				Filter: &lpb.RetrieveStudentPendingReallocateRequest_Filter{
					StartDate: timestamppb.New(time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)),
					EndDate:   timestamppb.New(time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC)),
				},
				Timezone: "UTC",
			})
			if err != nil {
				require.True(t, tc.hasErr)
			} else {
				require.False(t, tc.hasErr)
				require.NotNil(t, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestStudentSubscriptionGRPCService_GetStudentCoursesAndClasses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	studentSubscriptionRepo := new(mock_repositories.MockStudentSubscriptionRepo)
	studentId := "Student_id"

	testCases := []struct {
		name   string
		req    *lpb.GetStudentCoursesAndClassesRequest
		setup  func(context.Context)
		hasErr bool
		res    *lpb.GetStudentCoursesAndClassesResponse
	}{
		{
			name: "successfully",
			req: &lpb.GetStudentCoursesAndClassesRequest{
				StudentId: studentId,
			},
			setup: func(context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCoursesAndClasses", ctx, db, studentId).
					Return(&domain.StudentCoursesAndClasses{
						StudentID: studentId,
						Courses: []*domain.StudentCoursesAndClassesCourses{
							{
								CourseID: "course_id_0",
								Name:     "course_name_0",
							},
							{
								CourseID: "course_id_1",
								Name:     "course_name_1",
							},
						},
						Classes: []*domain.StudentCoursesAndClassesClasses{
							{
								ClassID:  "class_id_1",
								Name:     "class_name_1",
								CourseID: "course_id_1",
							},
						},
					}, nil).Once()
			},
			hasErr: false,
			res: &lpb.GetStudentCoursesAndClassesResponse{
				StudentId: studentId,
				Courses: []*cpb.Course{
					{
						Info: &cpb.ContentBasicInfo{
							Id:   "course_id_0",
							Name: "course_name_0",
						},
					},
					{
						Info: &cpb.ContentBasicInfo{
							Id:   "course_id_1",
							Name: "course_name_1",
						},
					},
				},
				Classes: []*lpb.GetStudentCoursesAndClassesResponse_Class{
					{
						ClassId:  "class_id_1",
						Name:     "class_name_1",
						CourseId: "course_id_1",
					},
				},
			},
		},
		{
			name: "there are no any class",
			req: &lpb.GetStudentCoursesAndClassesRequest{
				StudentId: studentId,
			},
			setup: func(context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCoursesAndClasses", ctx, db, studentId).
					Return(&domain.StudentCoursesAndClasses{
						StudentID: studentId,
						Courses: []*domain.StudentCoursesAndClassesCourses{
							{
								CourseID: "course_id_0",
								Name:     "course_name_0",
							},
							{
								CourseID: "course_id_1",
								Name:     "course_name_1",
							},
						},
						Classes: []*domain.StudentCoursesAndClassesClasses{},
					}, nil).Once()
			},
			hasErr: false,
			res: &lpb.GetStudentCoursesAndClassesResponse{
				StudentId: studentId,
				Courses: []*cpb.Course{
					{
						Info: &cpb.ContentBasicInfo{
							Id:   "course_id_0",
							Name: "course_name_0",
						},
					},
					{
						Info: &cpb.ContentBasicInfo{
							Id:   "course_id_1",
							Name: "course_name_1",
						},
					},
				},
			},
		},
		{
			name: "there are no any class and course",
			req: &lpb.GetStudentCoursesAndClassesRequest{
				StudentId: studentId,
			},
			setup: func(context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCoursesAndClasses", ctx, db, studentId).
					Return(&domain.StudentCoursesAndClasses{
						StudentID: studentId,
					}, nil).Once()
			},
			hasErr: false,
			res: &lpb.GetStudentCoursesAndClassesResponse{
				StudentId: studentId,
			},
		},
		{
			name: "there are no record",
			req: &lpb.GetStudentCoursesAndClassesRequest{
				StudentId: studentId,
			},
			setup: func(context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCoursesAndClasses", ctx, db, studentId).
					Return(nil, nil).Once()
			},
			hasErr: false,
			res:    nil,
		},
		{
			name: "got error",
			req: &lpb.GetStudentCoursesAndClassesRequest{
				StudentId: studentId,
			},
			setup: func(context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentSubscriptionRepo.
					On("GetStudentCoursesAndClasses", ctx, db, studentId).
					Return(nil, fmt.Errorf("error")).Once()
			},
			hasErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			src := NewStudentSubscriptionGRPCService(
				wrapperConnection,
				studentSubscriptionRepo,
				nil,
				nil,
				nil,
				"",
				nil,
			)
			res, err := src.GetStudentCoursesAndClasses(ctx, tc.req)
			if tc.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.res, res)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
