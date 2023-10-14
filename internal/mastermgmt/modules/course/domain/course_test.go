package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	course_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/mastermgmt/modules/course/infrastructure/repo"
	mock_location "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCourse_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	locationRepo := new(mock_location.MockLocationRepo)
	masterMgmtCourseTypeRepo := new(mock_repositories.MockCourseTypeRepo)

	tcs := []struct {
		name    string
		course  *domain.Course
		setup   func(ctx context.Context)
		isValid bool
	}{
		{
			name: "full fields",
			course: &domain.Course{
				CourseID:          "test-course-id-1",
				LocationID:        "test-location-id-1",
				Country:           "",
				Subject:           "",
				Grade:             0,
				DisplayOrder:      1,
				SchoolID:          1,
				TeacherIDs:        []string{},
				CourseType:        "test-course-type-1",
				Icon:              "test-icon-1",
				PresetStudyPlanID: "",
				Status:            "",
				StartDate:         now,
				EndDate:           now,
				TeachingMethod:    domain.CourseTeachingMethodGroup,
				Name:              "test-name-1",
				DeletedAt:         nil,
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			setup: func(ctx context.Context) {
				masterMgmtCourseTypeRepo.On("GetByIDs", ctx, db, []string{"test-course-type-1"}).Once().Return(
					[]*course_domain.CourseType{
						{
							CourseTypeID: "test-course-type-1",
							Name:         "name-1",
						},
					}, nil,
				)
			},
			isValid: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewCourse().
				WithLocationRepo(locationRepo).
				WithCourseTypeRepo(masterMgmtCourseTypeRepo).
				WithCourseID(tc.course.CourseID).
				WithLocationID(tc.course.LocationID).
				WithDisplayOrder(tc.course.DisplayOrder).
				WithSchoolID(tc.course.SchoolID).
				WithCourseType(tc.course.CourseType).
				WithIcon(tc.course.Icon).
				WithModificationTime(tc.course.CreatedAt, tc.course.UpdatedAt).
				WithName(tc.course.Name).
				WithTeachingMethod(tc.course.TeachingMethod)
			actual, err := builder.Build(ctx, db)
			if !tc.isValid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				assert.Equal(t, tc.course.CourseID, actual.CourseID)
				assert.Equal(t, tc.course.Name, actual.Name)
				assert.Equal(t, tc.course.CourseTypeID, actual.CourseType)
				assert.Equal(t, tc.course.TeachingMethod, actual.TeachingMethod)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				locationRepo,
				masterMgmtCourseTypeRepo,
			)
		})
	}
}
