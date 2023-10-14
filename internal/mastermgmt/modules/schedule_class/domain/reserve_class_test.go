package domain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_schedule_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/schedule_class/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReserveClass_Build(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_schedule_class_repo.MockReserveClassRepo)

	tcs := []struct {
		name         string
		reserveClass *domain.ReserveClass
		setup        func(ctx context.Context)
		expectedErr  error
	}{
		{
			name: "full fields",
			reserveClass: &domain.ReserveClass{
				ReserveClassID:   "reserve_class_id_01",
				StudentID:        "student_id_01",
				StudentPackageID: "student_package_id_01",
				CourseID:         "course_id_01",
				ClassID:          "class_id_01",
				EffectiveDate:    now,
				CreatedAt:        now,
				UpdatedAt:        now,
				Repo:             repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
		{
			name: "missing studentID",
			reserveClass: &domain.ReserveClass{
				ReserveClassID: "reserve_class_id_01",
				// StudentID:        "student_id_01",
				StudentPackageID: "student_package_id_01",
				CourseID:         "course_id_01",
				ClassID:          "class_id_01",
				EffectiveDate:    now,
				CreatedAt:        now,
				UpdatedAt:        now,
				Repo:             repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid reserve class: %w", fmt.Errorf("ReserveClass.StudentID cannot be empty")),
		},
		{
			name: "missing StudentPackageID",
			reserveClass: &domain.ReserveClass{
				ReserveClassID: "reserve_class_id_01",
				StudentID:      "student_id_01",
				// StudentPackageID: "student_package_id_01",
				CourseID:      "course_id_01",
				ClassID:       "class_id_01",
				EffectiveDate: now,
				CreatedAt:     now,
				UpdatedAt:     now,
				Repo:          repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid reserve class: %w", fmt.Errorf("ReserveClass.StudentPackageID cannot be empty")),
		},
		{
			name: "missing CourseID",
			reserveClass: &domain.ReserveClass{
				ReserveClassID:   "reserve_class_id_01",
				StudentID:        "student_id_01",
				StudentPackageID: "student_package_id_01",
				// CourseID:         "course_id_01",
				ClassID:       "class_id_01",
				EffectiveDate: now,
				CreatedAt:     now,
				UpdatedAt:     now,
				Repo:          repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid reserve class: %w", fmt.Errorf("ReserveClass.CourseID cannot be empty")),
		},
		{
			name: "missing ClassID",
			reserveClass: &domain.ReserveClass{
				ReserveClassID:   "reserve_class_id_01",
				StudentID:        "student_id_01",
				StudentPackageID: "student_package_id_01",
				CourseID:         "course_id_01",
				// ClassID:          "class_id_01",
				EffectiveDate: now,
				CreatedAt:     now,
				UpdatedAt:     now,
				Repo:          repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid reserve class: %w", fmt.Errorf("ReserveClass.ClassID cannot be empty")),
		},
		{
			name: "missing EffectiveDate",
			reserveClass: &domain.ReserveClass{
				ReserveClassID:   "reserve_class_id_01",
				StudentID:        "student_id_01",
				StudentPackageID: "student_package_id_01",
				CourseID:         "course_id_01",
				ClassID:          "class_id_01",
				// EffectiveDate:    now,
				CreatedAt:        now,
				UpdatedAt:        now,
				Repo:             repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid reserve class: %w", fmt.Errorf("ReserveClass.EffectiveDate cannot be empty")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewReserveClassBuilder().
				WithReserveClassID(tc.reserveClass.ReserveClassID).
				WithReserveClassRepo(repo).WithStudentID(tc.reserveClass.StudentID).
				WithStudentPackageID(tc.reserveClass.StudentPackageID).
				WithCourseID(tc.reserveClass.CourseID).
				WithClassID(tc.reserveClass.ClassID).
				WithEffectiveDate(tc.reserveClass.EffectiveDate).
				WithModificationTime(tc.reserveClass.CreatedAt, tc.reserveClass.UpdatedAt)

			actual, err := builder.Build()
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.reserveClass, actual)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				repo,
				repo,
			)
		})
	}
}
