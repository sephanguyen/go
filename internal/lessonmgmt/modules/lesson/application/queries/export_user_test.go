package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	master_data_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	master_data_repo "github.com/manabie-com/backend/mock/lessonmgmt/master_data/repositories"
	user_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExportUserHandler_ExportTeacher(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	teacherRepo := new(user_repositories.MockTeacherRepo)
	userBasicInfoRepo := new(user_repositories.MockUserBasicInfoRepo)
	locationRepo := new(master_data_repo.MockLocationRepository)
	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				teacherRepo.On("ListByGrantedLocation", ctx, mock.Anything).Return(map[string][]string{
					"teacher-id1": {"location-id1", "location-id2"},
					"teacher-id2": {"location-id1"},
					"teacher-id3": {"location-id3"},
				}, nil).Once()
				userBasicInfoRepo.On("GetUser", ctx, mock.Anything, mock.Anything).Return([]*repo.UserBasicInfo{
					{
						UserID:   database.Text("teacher-id1"),
						FullName: database.Text("teacher1"),
					},
					{
						UserID:   database.Text("teacher-id2"),
						FullName: database.Text("teacher2"),
					},
					{
						UserID:   database.Text("teacher-id3"),
						FullName: database.Text("teacher3"),
					},
				}, nil).Once()
				locationRepo.On("GetLocationByID", ctx, mock.Anything, mock.Anything).Return([]*master_data_domain.Location{
					{
						LocationID:        database.Text("location-id1"),
						PartnerInternalID: database.Text("pi1"),
						Name:              database.Text("location1"),
					},
					{
						LocationID:        database.Text("location-id2"),
						PartnerInternalID: database.Text("pi2"),
						Name:              database.Text("location2"),
					},
					{
						LocationID:        database.Text("location-id3"),
						PartnerInternalID: database.Text("pi3"),
						Name:              database.Text("location3"),
					},
				}, nil).Once()
			},
			hasError: false,
		},
	}
	exportTeacherHandler := ExportUserHandler{
		WrapperConnection: wrapperConnection,
		TeacherRepo:       teacherRepo,
		UserBasicInfoRepo: userBasicInfoRepo,
		LocationRepo:      locationRepo,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			data, err := exportTeacherHandler.ExportTeacher(ctx)
			if err != nil {
				require.True(t, tc.hasError)
			}
			fmt.Println(data)
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestExportUserHandler_ExportEnrolledStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	userBasicInfoRepo := new(user_repositories.MockUserBasicInfoRepo)
	locationRepo := new(master_data_repo.MockLocationRepository)
	studentSubRepo := new(user_repositories.MockStudentSubscriptionRepo)
	courseRepo := new(master_data_repo.MockCourseRepository)
	timezone := "UTC"
	loc, _ := time.LoadLocation(timezone)

	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				studentSubRepo.On("GetAll", ctx, mock.Anything).Return([]*user_domain.EnrolledStudent{
					{
						StudentID:        "st1",
						EnrollmentStatus: string(domain.EnrollmentStatusEnrolled),
						CourseID:         "c1",
						LocationID:       "l1",
						StartAt:          time.Date(2022, 01, 30, 0, 0, 0, 0, loc),
						EndAt:            time.Date(2022, 02, 28, 0, 0, 0, 0, loc),
					},
					{
						StudentID:        "st2",
						EnrollmentStatus: string(domain.EnrollmentStatusEnrolled),
						CourseID:         "c2",
						LocationID:       "l2",
						StartAt:          time.Date(2022, 01, 30, 0, 0, 0, 0, loc),
						EndAt:            time.Date(2022, 02, 28, 0, 0, 0, 0, loc),
					},
				}, nil).Once()
				userBasicInfoRepo.On("GetUser", ctx, mock.Anything, mock.Anything).Return([]*repo.UserBasicInfo{
					{
						UserID:   database.Text("st1"),
						FullName: database.Text("student1"),
					},
					{
						UserID:   database.Text("st2"),
						FullName: database.Text("student2"),
					},
				}, nil).Once()
				locationRepo.On("GetLocationByID", ctx, mock.Anything, mock.Anything).Return([]*master_data_domain.Location{
					{
						LocationID:        database.Text("l1"),
						PartnerInternalID: database.Text("pi1"),
						Name:              database.Text("location1"),
					},
					{
						LocationID:        database.Text("l2"),
						PartnerInternalID: database.Text("pi2"),
						Name:              database.Text("location2"),
					},
				}, nil).Once()
				courseRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]*master_data_domain.Course{
					{
						CourseID: database.Text("c1"),
						Name:     database.Text("course1"),
					},
					{
						CourseID: database.Text("c2"),
						Name:     database.Text("course2"),
					},
				}, nil).Once()
			},
			hasError: false,
		},
	}
	exportUserHandler := ExportUserHandler{
		WrapperConnection:       wrapperConnection,
		UserBasicInfoRepo:       userBasicInfoRepo,
		LocationRepo:            locationRepo,
		CourseRepo:              courseRepo,
		StudentSubscriptionRepo: studentSubRepo,
		UnleashClient:           mockUnleashClient,
	}
	data := [][]string{
		{"student_id", "student_name", "student_status", "partner_internal_id", "granted_location_id", "location_name", "course_id", "course_name", "course_start_date", "course_end_date"},
		{"st1", "student1", "Enrolled", "pi1", "l1", "location1", "c1", "course1", "2022/01/30", "2022/02/28"},
		{"st2", "student2", "Enrolled", "pi2", "l2", "location2", "c2", "course2", "2022/01/30", "2022/02/28"},
	}
	expectedData := exporter.ToCSV(data)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			data, err := exportUserHandler.ExportEnrolledStudent(ctx, timezone)
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
				require.Equal(t, expectedData, data)
			}
		})
	}
}
