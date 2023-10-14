package controller

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/assigned_student/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestAssignedStudentGRPCService_GetAssignedStudentList(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	assignedStudentRepo := &mock_repositories.MockAssignedStudentRepo{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	s := &AssignedStudentGRPCService{
		QueryHandler: queries.AssignedStudentQueryHandler{
			WrapperConnection:   wrapperConnection,
			AssignedStudentRepo: assignedStudentRepo,
		},
		env:              "local",
		unleashClientIns: mockUnleashClient,
	}

	courseId := "course-1"
	courses := []string{courseId}
	students := []string{"student-1"}
	locations := []string{"center-1", "center-2"}
	timezone := "timezone-1"

	testCases := []TestCase{
		{
			name: "School Admin get list assigned student slot successfully with filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.GetAssignedStudentListRequest{
				PurchaseMethod: lpb.PurchaseMethod_PURCHASE_METHOD_SLOT,
				Paging:         &cpb.Paging{Limit: 2},
				Keyword:        "student name",
				Filter: &lpb.GetAssignedStudentListRequest_Filter{
					LocationIds: []string{"center-1", "center-2"},
					CourseIds:   []string{"course-1"},
					StudentIds:  []string{"student-1"},
				},
				LocationIds: locations,
				Timezone:    timezone,
			},
			expectedErr: nil,
			expectedResp: &lpb.GetAssignedStudentListResponse{
				Items: []*lpb.AssignedStudentInfo{
					{
						StudentId:     "student-1",
						CourseId:      courseId,
						Duration:      mock.Anything,
						LocationId:    "center-1",
						PurchasedSlot: int32(4),
						AssignedSlot:  int32(2),
						SlotGap:       int32(-2),
						Status:        lpb.AssignedStudentStatus_STUDENT_STATUS_UNDER_ASSIGNED,
					},
					{
						StudentId:     "student-2",
						CourseId:      courseId,
						Duration:      mock.Anything,
						LocationId:    "center-2",
						PurchasedSlot: int32(4),
						AssignedSlot:  int32(2),
						SlotGap:       int32(-2),
						Status:        lpb.AssignedStudentStatus_STUDENT_STATUS_UNDER_ASSIGNED,
					},
				},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "student-sub-id-2",
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
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				assignedStudentRepo.On("GetAssignedStudentList", mock.Anything, db, &payloads.GetAssignedStudentListArg{
					Limit:                     2,
					CourseIDs:                 courses,
					StudentIDs:                students,
					KeyWord:                   "student name",
					LocationIDs:               locations,
					PurchaseMethod:            string(domain.PurchaseMethodSlot),
					Timezone:                  timezone,
				}).Once().Return([]*domain.AssignedStudent{
					{
						StudentID:             "student-1",
						CourseID:              courseId,
						LocationID:            "center-1",
						Duration:              mock.Anything,
						PurchasedSlot:         int32(4),
						AssignedSlot:          int32(2),
						SlotGap:               int32(-2),
						AssignedStatus:        domain.AssignedStudentStatusUnderAssigned,
						StudentSubscriptionID: "student-sub-id-1",
					},
					{
						StudentID:             "student-2",
						CourseID:              courseId,
						LocationID:            "center-2",
						Duration:              mock.Anything,
						PurchasedSlot:         int32(4),
						AssignedSlot:          int32(2),
						SlotGap:               int32(-2),
						AssignedStatus:        domain.AssignedStudentStatusUnderAssigned,
						StudentSubscriptionID: "student-sub-id-2",
					},
				}, uint32(99), "", uint32(99), nil)
			},
		},
		{
			name: "School Admin get list assigned student slot successfully without filter",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.GetAssignedStudentListRequest{
				PurchaseMethod: lpb.PurchaseMethod_PURCHASE_METHOD_SLOT,
				Paging:         &cpb.Paging{Limit: 2},
				Keyword:        "",
				LocationIds:    locations,
				Timezone:       timezone,
			},
			expectedErr: nil,
			expectedResp: &lpb.GetAssignedStudentListResponse{
				Items: []*lpb.AssignedStudentInfo{
					{
						StudentId:     "student-1",
						CourseId:      courseId,
						Duration:      mock.Anything,
						LocationId:    "center-1",
						PurchasedSlot: int32(2),
						AssignedSlot:  int32(4),
						SlotGap:       int32(2),
						Status:        lpb.AssignedStudentStatus_STUDENT_STATUS_OVER_ASSIGNED,
					},
					{
						StudentId:     "student-2",
						CourseId:      courseId,
						Duration:      mock.Anything,
						LocationId:    "center-2",
						PurchasedSlot: int32(2),
						AssignedSlot:  int32(2),
						SlotGap:       int32(0),
						Status:        lpb.AssignedStudentStatus_STUDENT_STATUS_JUST_ASSIGNED,
					},
				},
				NextPage: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetString{
						OffsetString: "student-sub-id-2",
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
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				assignedStudentRepo.On("GetAssignedStudentList", mock.Anything, db, &payloads.GetAssignedStudentListArg{
					Limit:                     2,
					LocationIDs:               locations,
					PurchaseMethod:            string(domain.PurchaseMethodSlot),
					Timezone:                  timezone,
				}).Once().Return([]*domain.AssignedStudent{
					{
						StudentID:             "student-1",
						CourseID:              courseId,
						LocationID:            "center-1",
						Duration:              mock.Anything,
						PurchasedSlot:         int32(2),
						AssignedSlot:          int32(4),
						SlotGap:               int32(2),
						AssignedStatus:        OverAssignedStatus,
						StudentSubscriptionID: "student-sub-id-1",
					},
					{
						StudentID:             "student-2",
						CourseID:              courseId,
						LocationID:            "center-2",
						Duration:              mock.Anything,
						PurchasedSlot:         int32(2),
						AssignedSlot:          int32(2),
						SlotGap:               int32(0),
						AssignedStatus:        JustAssignedStatus,
						StudentSubscriptionID: "student-sub-id-2",
					},
				}, uint32(99), "", uint32(99), nil)
			},
		},
		{
			name: "Return list empty successfully",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.GetAssignedStudentListRequest{
				PurchaseMethod: lpb.PurchaseMethod_PURCHASE_METHOD_SLOT,
				Paging:         &cpb.Paging{Limit: 2},
				Timezone:       timezone,
			},
			expectedErr: nil,
			expectedResp: &lpb.GetAssignedStudentListResponse{
				Items: []*lpb.AssignedStudentInfo{},
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
				assignedStudentRepo.On("GetAssignedStudentList", mock.Anything, db, &payloads.GetAssignedStudentListArg{
					Limit:                     2,
					PurchaseMethod:            string(domain.PurchaseMethodSlot),
					Timezone:                  timezone,
				}).Once().Return([]*domain.AssignedStudent{}, uint32(0), "", uint32(0), nil)
			},
		},
		{
			name: "Return fail missing page",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &lpb.GetAssignedStudentListRequest{
				PurchaseMethod: lpb.PurchaseMethod_PURCHASE_METHOD_SLOT,
				Keyword:        "",
				LocationIds:    locations,
				Timezone:       timezone,
			},
			expectedErr:  status.Error(codes.Internal, "missing paging info"),
			expectedResp: &lpb.GetAssignedStudentListResponse{},
			setup:        func(ctx context.Context) {},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*lpb.GetAssignedStudentListRequest)
			resp, err := s.GetAssignedStudentList(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
