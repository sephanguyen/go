package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	masterdata_repo "github.com/manabie-com/backend/mock/lessonmgmt/master_data/repositories"
	user_repo "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	"github.com/stretchr/testify/require"
)

func TestGetByStudentSubscription(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockCourseRepo := &masterdata_repo.MockCourseRepository{}
	mockClassRepo := &masterdata_repo.MockClassRepository{}
	mockStudentSubscriptionRepo := &user_repo.MockStudentSubscriptionRepo{}
	db := &mock_database.Ext{}
	testCases := []struct {
		name   string
		input  []string
		output []*domain.ClassUnassigned
		setup  func(ctx context.Context)
		hasErr bool
	}{
		{
			name:  "happy case",
			input: []string{"ss1", "ss2", "ss3"},
			output: []*domain.ClassUnassigned{
				{
					StudentSubscriptionID: "ss1",
					IsClassUnAssigned:     false,
				},
				{
					StudentSubscriptionID: "ss2",
					IsClassUnAssigned:     false,
				},
				{
					StudentSubscriptionID: "ss3",
					IsClassUnAssigned:     true,
				},
			},
			setup: func(ctx context.Context) {
				mockStudentSubscriptionRepo.On("GetByStudentSubscriptionIDs", ctx, db, []string{"ss1", "ss2", "ss3"}).
					Return([]*user_domain.StudentSubscription{
						{
							StudentSubscriptionID: "ss1",
							StudentID:             "s1",
							CourseID:              "c1",
						},
						{
							StudentSubscriptionID: "ss2",
							StudentID:             "s2",
							CourseID:              "c2",
						},
						{
							StudentSubscriptionID: "ss3",
							StudentID:             "s3",
							CourseID:              "c3",
						},
					}, nil).Once()
				mockCourseRepo.On("GetByIDs", ctx, db, []string{"c1", "c2", "c3"}).
					Return([]*domain.Course{
						{
							CourseID:       database.Text("c1"),
							TeachingMethod: database.Text("COURSE_TEACHING_METHOD_INDIVIDUAL"),
						},
						{
							CourseID:       database.Text("c2"),
							TeachingMethod: database.Text("COURSE_TEACHING_METHOD_GROUP"),
						},
						{
							CourseID:       database.Text("c3"),
							TeachingMethod: database.Text("COURSE_TEACHING_METHOD_GROUP"),
						},
					}, nil).Once()
				mockClassRepo.On("GetByStudentCourse", ctx, db, []string{"s2", "c2", "s3", "c3"}).
					Return(map[string]string{
						"s2-c2": "class1",
					}, nil).Once()
				mockClassRepo.On("GetReserveClass", ctx, db, []string{"s2", "c2", "s3", "c3"}).
					Return(map[string]string{}, nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		tc.setup(ctx)
		t.Run(tc.name, func(t *testing.T) {
			classUCase := &ClassUseCase{
				ClassRepo:               mockClassRepo,
				CourseRepo:              mockCourseRepo,
				StudentSubscriptionRepo: mockStudentSubscriptionRepo,
			}
			resp, err := classUCase.GetByStudentSubscription(ctx, db, tc.input)
			if err != nil {
				require.True(t, tc.hasErr)
			} else {
				require.False(t, tc.hasErr)
				fmt.Println(len(resp))
				for _, ss := range resp {
					fmt.Println(ss.StudentSubscriptionID)
					fmt.Println(ss.IsClassUnAssigned)
				}
				require.Equal(t, tc.output, resp)
			}
		})
	}
}
