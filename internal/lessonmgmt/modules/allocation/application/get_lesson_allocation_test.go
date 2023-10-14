package application

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/domain"
	masterdata_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_allocation/repositories"
	academic_year_repositories "github.com/manabie-com/backend/mock/lessonmgmt/master_data/repositories"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetLessonAllocationHandler_GetLessonAllocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	now := time.Now()
	lessonAllocationRepo := new(mock_repositories.MockLessonAllocationRepo)
	academicYeerRepo := new(academic_year_repositories.MockAcademicYearRepository)

	testCase := []struct {
		name     string
		setup    func(context.Context)
		req      *GetLessonAllocationRequest
		resp     *GetLessonAllocationResponse
		hasError bool
	}{
		{
			name: "without filter",
			req: &GetLessonAllocationRequest{
				Paging: support.Paging[int]{
					Limit:  3,
					Offset: 0,
				},
			},
			resp: &GetLessonAllocationResponse{
				Total:                    39,
				TotalOfNoneAssigned:      12,
				TotalOfPartiallyAssigned: 13,
				TotalOfFullyAssigned:     14,
				TotalOfOverAssigned:      0,
				Items: []Item{
					{
						StudentSubscriptionID: "subscription_id1",
						StudentID:             "student-1",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          0,
						AllocationStatus:      "NONE_ASSIGNED",
					},
					{
						StudentSubscriptionID: "subscription_id2",
						StudentID:             "student-2",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          5,
						AllocationStatus:      "PARTIALLY_ASSIGNED",
					},
					{
						StudentSubscriptionID: "subscription_id3",
						StudentID:             "student-3",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          10,
						AllocationStatus:      "FULLY_ASSIGNED",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				academicYeerRepo.On("GetCurrentAcademicYear", mock.Anything, mock.Anything).Once().Return(&masterdata_domain.AcademicYear{
					AcademicYearID: database.Text("y1"),
				}, nil)
				lessonAllocationRepo.On("GetLessonAllocation", mock.Anything, mock.Anything, domain.LessonAllocationFilter{
					Limit:          3,
					Offset:         0,
					TeachingMethod: []domain.CourseTeachingMethod{},
				}).Once().Return([]*domain.AllocatedStudent{
					{
						StudentSubscriptionID: "subscription_id1",
						StudentID:             "student-1",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          0,
					},
					{
						StudentSubscriptionID: "subscription_id2",
						StudentID:             "student-2",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          5,
					},
					{
						StudentSubscriptionID: "subscription_id3",
						StudentID:             "student-3",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          10,
					},
				}, map[string]uint32{
					"NONE_ASSIGNED":      12,
					"PARTIALLY_ASSIGNED": 13,
					"FULLY_ASSIGNED":     14,
					"OVER_ASSIGNED":      0,
				}, nil)
			},
		},
		{
			name: "with filter",
			req: &GetLessonAllocationRequest{
				Filter: struct {
					CourseID               []string
					CourseTypeID           []string
					LocationID             []string
					TeachingMethod         []string
					StartDate              time.Time
					EndDate                time.Time
					IsOnlyReallocation     bool
					IsClassUnassigned      bool
					LessonAllocationStatus string
					ProductID              []string
				}{
					[]string{"course-1"},
					[]string{"course-type-1"},
					[]string{"location-1"},
					[]string{},
					now.AddDate(0, 0, -14),
					now.AddDate(0, 0, 14),
					false,
					false,
					string(domain.PartiallyAssigned),
					[]string{},
				},
				LocationSettings: []string{"location-1", "location-2"},
				Paging: support.Paging[int]{
					Limit:  3,
					Offset: 0,
				},
			},
			resp: &GetLessonAllocationResponse{
				Total:                    39,
				TotalOfNoneAssigned:      12,
				TotalOfPartiallyAssigned: 13,
				TotalOfFullyAssigned:     14,
				TotalOfOverAssigned:      0,
				Items: []Item{
					{
						StudentSubscriptionID: "subscription_id1",
						StudentID:             "student-1",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          5,
						AllocationStatus:      "PARTIALLY_ASSIGNED",
					},
					{
						StudentSubscriptionID: "subscription_id2",
						StudentID:             "student-2",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          4,
						AllocationStatus:      "PARTIALLY_ASSIGNED",
					},
					{
						StudentSubscriptionID: "subscription_id3",
						StudentID:             "student-3",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          3,
						AllocationStatus:      "PARTIALLY_ASSIGNED",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				academicYeerRepo.On("GetCurrentAcademicYear", mock.Anything, mock.Anything).Once().Return(&masterdata_domain.AcademicYear{
					AcademicYearID: database.Text("y1"),
					StartDate:      database.DateFromPb(timestamppb.New(now.AddDate(0, -1, 0))),
					EndDate:        database.DateFromPb(timestamppb.New(now.AddDate(0, 1, 0))),
				}, nil)
				lessonAllocationRepo.On("GetLessonAllocation", mock.Anything, mock.Anything, domain.LessonAllocationFilter{
					Limit:                  3,
					Offset:                 0,
					CourseID:               []string{"course-1"},
					CourseTypeID:           []string{"course-type-1"},
					LocationID:             []string{"location-1"},
					TeachingMethod:         []domain.CourseTeachingMethod{},
					StartDate:              now.AddDate(0, 0, -14),
					EndDate:                now.AddDate(0, 0, 14),
					IsOnlyReallocation:     false,
					LessonAllocationStatus: domain.PartiallyAssigned,
					ProductID:              []string{},
				}).Once().Return([]*domain.AllocatedStudent{
					{
						StudentSubscriptionID: "subscription_id1",
						StudentID:             "student-1",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          5,
					},
					{
						StudentSubscriptionID: "subscription_id2",
						StudentID:             "student-2",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          4,
					},
					{
						StudentSubscriptionID: "subscription_id3",
						StudentID:             "student-3",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          3,
					},
				}, map[string]uint32{
					"NONE_ASSIGNED":      0,
					"PARTIALLY_ASSIGNED": 13,
					"FULLY_ASSIGNED":     0,
					"OVER_ASSIGNED":      0,
				}, nil)
			},
		},
		{
			name: "get by academic year",
			req: &GetLessonAllocationRequest{
				Filter: struct {
					CourseID               []string
					CourseTypeID           []string
					LocationID             []string
					TeachingMethod         []string
					StartDate              time.Time
					EndDate                time.Time
					IsOnlyReallocation     bool
					IsClassUnassigned      bool
					LessonAllocationStatus string
					ProductID              []string
				}{
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					time.Date(1970, 0, 0, 0, 0, 0, 0, time.UTC),
					time.Date(1970, 0, 0, 0, 0, 0, 0, time.UTC),
					false,
					false,
					string(domain.All),
					[]string{},
				},
				LocationSettings: []string{"location-1", "location-2"},
				Paging: support.Paging[int]{
					Limit:  3,
					Offset: 0,
				},
			},
			resp: &GetLessonAllocationResponse{
				Total:                    39,
				TotalOfNoneAssigned:      12,
				TotalOfPartiallyAssigned: 13,
				TotalOfFullyAssigned:     14,
				TotalOfOverAssigned:      0,
				Items: []Item{
					{
						StudentSubscriptionID: "subscription_id1",
						StudentID:             "student-1",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          5,
						AllocationStatus:      "PARTIALLY_ASSIGNED",
					},
					{
						StudentSubscriptionID: "subscription_id2",
						StudentID:             "student-2",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          4,
						AllocationStatus:      "PARTIALLY_ASSIGNED",
					},
					{
						StudentSubscriptionID: "subscription_id3",
						StudentID:             "student-3",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          3,
						AllocationStatus:      "PARTIALLY_ASSIGNED",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				academicYeerRepo.On("GetCurrentAcademicYear", mock.Anything, mock.Anything).Once().Return(&masterdata_domain.AcademicYear{
					AcademicYearID: database.Text("y1"),
					StartDate:      database.DateFromPb(timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))),
					EndDate:        database.DateFromPb(timestamppb.New(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC))),
				}, nil)
				lessonAllocationRepo.On("GetLessonAllocation", mock.Anything, mock.Anything, domain.LessonAllocationFilter{
					Limit:                  3,
					Offset:                 0,
					CourseID:               []string{},
					CourseTypeID:           []string{},
					LocationID:             []string{"location-1", "location-2"},
					TeachingMethod:         []domain.CourseTeachingMethod{},
					StartDate:              time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:                time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
					IsOnlyReallocation:     false,
					LessonAllocationStatus: domain.All,
					ProductID:              []string{},
				}).Once().Return([]*domain.AllocatedStudent{
					{
						StudentSubscriptionID: "subscription_id1",
						StudentID:             "student-1",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          5,
					},
					{
						StudentSubscriptionID: "subscription_id2",
						StudentID:             "student-2",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          4,
					},
					{
						StudentSubscriptionID: "subscription_id3",
						StudentID:             "student-3",
						CourseID:              "course-1",
						LocationID:            "location-1",
						StartTime:             now.AddDate(0, -1, 0),
						EndTime:               now.AddDate(0, 1, 0),
						PurchasedSlot:         10,
						AssignedSlot:          3,
					},
				}, map[string]uint32{
					"NONE_ASSIGNED":      0,
					"PARTIALLY_ASSIGNED": 13,
					"FULLY_ASSIGNED":     0,
					"OVER_ASSIGNED":      0,
				}, nil)
			},
		},
	}
	handler := GetLessonAllocationHandler{
		LessonAllocationRepo: lessonAllocationRepo,
		AcademicYearRepo:     academicYeerRepo,
		WrapperConnection:    wrapperConnection,
	}
	for _, tc := range testCase {
		tc.setup(ctx)
		t.Run(tc.name, func(t *testing.T) {
			res, err := handler.GetLessonAllocation(ctx, tc.req)
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				res.Total = tc.resp.Total
				res.TotalOfFullyAssigned = tc.resp.TotalOfFullyAssigned
				res.TotalOfPartiallyAssigned = tc.resp.TotalOfPartiallyAssigned
				res.TotalOfNoneAssigned = tc.resp.TotalOfNoneAssigned
				res.TotalOfOverAssigned = tc.resp.TotalOfOverAssigned
				require.EqualValues(t, tc.resp.Items, res.Items)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
