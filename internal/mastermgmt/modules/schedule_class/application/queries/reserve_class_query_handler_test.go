package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_schedule_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/schedule_class/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupMock() (*ReserveClassQueryHandler, *mock_database.Ext, *mock_schedule_class_repo.MockReserveClassRepo, *mock_schedule_class_repo.MockCourseRepo, *mock_schedule_class_repo.MockClassRepo, *mock_schedule_class_repo.MockStudentPackageClassRepo) {
	mockDB := &mock_database.Ext{}
	mockReserveClassRepo := new(mock_schedule_class_repo.MockReserveClassRepo)
	mockCourseRepo := new(mock_schedule_class_repo.MockCourseRepo)
	mockClassRepo := new(mock_schedule_class_repo.MockClassRepo)
	mockStudentPackageRepo := new(mock_schedule_class_repo.MockStudentPackageClassRepo)

	r := &ReserveClassQueryHandler{
		DB:                      mockDB,
		ReserveClassRepo:        mockReserveClassRepo,
		CourseRepo:              mockCourseRepo,
		ClassRepo:               mockClassRepo,
		StudentPackageClassRepo: mockStudentPackageRepo,
	}

	return r, mockDB, mockReserveClassRepo, mockCourseRepo, mockClassRepo, mockStudentPackageRepo
}

func TestReserveClassQueryHandler_RetrieveScheduledClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB, mockReserveClassRepo, mockCourseRepo, mockClassRepo, mockStudentPackageRepo := setupMock()
	studentID := "student_id"
	courseID := "course_id"
	scheduledClassID := "scheduled_class_id"
	studentPackageID := "student_package_id"
	currentClassID := "current_class_id"
	spcID := "scheduled_class_id-student_id-course_id"
	effectiveDate := time.Now().Add(30 * 24 * time.Hour)

	reserveClasses := []*domain.ReserveClass{
		{
			ReserveClassID:   "reserve_class_id",
			StudentID:        studentID,
			StudentPackageID: studentPackageID,
			CourseID:         courseID,
			ClassID:          scheduledClassID,
			EffectiveDate:    effectiveDate,
		},
	}

	spcItem := &repo.StudentPackageClassDTO{
		StudentPackageID: database.Text(studentPackageID),
		StudentID:        database.Text(studentID),
		CourseID:         database.Text(courseID),
		ClassID:          database.Text(currentClassID),
	}

	spcList := []*repo.StudentPackageClassDTO{
		spcItem,
	}

	mapSpc := map[string]*repo.StudentPackageClassDTO{
		spcID: spcItem,
	}

	mapCourse := map[string]*repo.Course{
		courseID: {
			CourseID: database.Text(courseID),
			Name:     database.Text("course_name"),
		},
	}

	mapClass := map[string]*repo.Class{
		currentClassID: {
			ClassID: database.Text(currentClassID),
			Name:    database.Text("Class A"),
		},
		scheduledClassID: {
			ClassID: database.Text(scheduledClassID),
			Name:    database.Text("Class B"),
		},
	}

	testCases := []struct {
		name         string
		expectedErr  error
		expectedResp *mpb.RetrieveScheduledStudentClassResponse
		setup        func(ctx context.Context)
	}{
		{
			name:         "GetByStudentIDs fail",
			expectedErr:  fmt.Errorf("query reserve class fail: error"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:        "GetByStudentIDs return empty",
			expectedErr: nil,
			expectedResp: &mpb.RetrieveScheduledStudentClassResponse{
				ScheduledClasses: []*mpb.RetrieveScheduledStudentClassResponse_ScheduledClassInfo{},
			},
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(nil, nil)
			},
		},
		{
			name:         "GetManyByStudentPackageIDAndStudentIDAndCourseID fail",
			expectedErr:  fmt.Errorf("query student package class fail: error"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(nil, map[string]*repo.StudentPackageClassDTO{}, fmt.Errorf("error"))
			},
		},
		{
			name:         "GetMapCourseByIDs fail",
			expectedErr:  fmt.Errorf("query course fail: error"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(spcList, mapSpc, nil)
				mockCourseRepo.On("GetMapCourseByIDs", ctx, mockDB, []string{courseID}).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "GetMapClassByIDs fail",
			expectedErr:  fmt.Errorf("query class fail: error"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(spcList, mapSpc, nil)
				mockCourseRepo.On("GetMapCourseByIDs", ctx, mockDB, []string{courseID}).Once().Return(mapCourse, nil)
				mockClassRepo.On("GetMapClassByIDs", ctx, mockDB, []string{currentClassID, scheduledClassID}).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "success",
			expectedErr:  nil,
			expectedResp: &mpb.RetrieveScheduledStudentClassResponse{
				ScheduledClasses: []*mpb.RetrieveScheduledStudentClassResponse_ScheduledClassInfo{
					{
						CourseId:   courseID,
						CourseName: "course_name",
						CurrentClass: &mpb.RetrieveScheduledStudentClassResponse_ClassInfo{
							ClassId: currentClassID,
							Name:    "Class A",
						},
						ScheduledClass: &mpb.RetrieveScheduledStudentClassResponse_ClassInfo{
							ClassId: scheduledClassID,
							Name:    "Class B",
						},
						EffectiveDate: &timestamppb.Timestamp{Seconds: effectiveDate.Unix()},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(spcList, mapSpc, nil)
				mockCourseRepo.On("GetMapCourseByIDs", ctx, mockDB, []string{courseID}).Once().Return(mapCourse, nil)
				mockClassRepo.On("GetMapClassByIDs", ctx, mockDB, []string{currentClassID, scheduledClassID}).Once().Return(mapClass, nil)
				mockStudentPackageRepo.On("GetStudentPackageClassID", studentPackageID, studentID, courseID).Once().Return(spcID)
			},
		},
		{
			name:         "cannot find course from map",
			expectedErr:  fmt.Errorf("not found course"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(spcList, mapSpc, nil)
				mockCourseRepo.On("GetMapCourseByIDs", ctx, mockDB, []string{courseID}).Once().Return(map[string]*repo.Course{}, nil)
				mockClassRepo.On("GetMapClassByIDs", ctx, mockDB, []string{currentClassID, scheduledClassID}).Once().Return(mapClass, nil)
				mockStudentPackageRepo.On("GetStudentPackageClassID", studentPackageID, studentID, courseID).Once().Return(spcID)
			},
		},
		{
			name:         "cannot find scheduled class from map",
			expectedErr:  fmt.Errorf("not found scheduled class"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(spcList, mapSpc, nil)
				mockCourseRepo.On("GetMapCourseByIDs", ctx, mockDB, []string{courseID}).Once().Return(mapCourse, nil)
				mockClassRepo.On("GetMapClassByIDs", ctx, mockDB, []string{currentClassID, scheduledClassID}).Once().Return(map[string]*repo.Class{}, nil)
				mockStudentPackageRepo.On("GetStudentPackageClassID", studentPackageID, studentID, courseID).Once().Return(spcID)
			},
		},
		{
			name:         "cannot find student package class from map",
			expectedErr:  fmt.Errorf("not found student package class"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(spcList, map[string]*repo.StudentPackageClassDTO{}, nil)
				mockCourseRepo.On("GetMapCourseByIDs", ctx, mockDB, []string{courseID}).Once().Return(mapCourse, nil)
				mockClassRepo.On("GetMapClassByIDs", ctx, mockDB, []string{currentClassID, scheduledClassID}).Once().Return(mapClass, nil)
				mockStudentPackageRepo.On("GetStudentPackageClassID", studentPackageID, studentID, courseID).Once().Return(spcID)
			},
		},
		{
			name:        "cannot find current class from map",
			expectedErr: fmt.Errorf("not found current active class"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				mockReserveClassRepo.On("GetByStudentIDs", ctx, mockDB, studentID).Once().Return(reserveClasses, nil)
				mockStudentPackageRepo.On("GetManyByStudentPackageIDAndStudentIDAndCourseID", ctx, mockDB, mock.Anything).Once().Return(spcList, mapSpc, nil)
				mockCourseRepo.On("GetMapCourseByIDs", ctx, mockDB, []string{courseID}).Once().Return(mapCourse, nil)
				mockClassRepo.On("GetMapClassByIDs", ctx, mockDB, []string{currentClassID, scheduledClassID}).Once().Return(map[string]*repo.Class{
					scheduledClassID: {
						ClassID: database.Text(scheduledClassID),
						Name:    database.Text("Class A"),
					},
				}, nil)
				mockStudentPackageRepo.On("GetStudentPackageClassID", studentPackageID, studentID, courseID).Once().Return(spcID)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := r.RetrieveScheduledClass(ctx, studentID)
			if tc.expectedErr != nil {
				assert.EqualError(t, err, tc.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				// assert.Equal(t, tc.expectedResp.ScheduledClasses, resp.ScheduledClasses)
			}

			mock.AssertExpectationsForObjects(t, mockReserveClassRepo)
		})
	}
}
