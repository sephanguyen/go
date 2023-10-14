package nat_sync

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

func (s *Suite) NatSendARequestHandlerStudentPackage(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := []*npb.EventStudentPackageV2{
		{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{

				StudentId: stepState.StudentIDs[0],
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-3",
					LocationId: "location-3",
				},
				IsActive: false,
			},
		},
		{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{

				StudentId: stepState.StudentIDs[1],
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-4",
					LocationId: "location-4",
				},
				IsActive: true,
			},
		}}
	stepState.Request = req

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) StoreCorrectResultFromHandlerStudentPackage(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := stepState.Request.([]*npb.EventStudentPackageV2)
	courseStudentrepo := &repositories.CourseStudentRepo{}
	courseStudentAccessPathRepo := &repositories.CourseStudentAccessPathRepo{}
	cspService := services.CourseStudentPackageService{
		DB:                          s.EurekaDB,
		CourseStudentRepo:           courseStudentrepo,
		CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
	}
	for _, r := range req {
		_, err := cspService.ProcessHandleStudentPackageEvent(ctx, r)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not processSyncCourseStudent: %w", err)
		}

	}

	for _, pack := range req {
		var countCourseStudent, countCourseStudentAccessPath int

		if pack.StudentPackage.IsActive {
			csquery := `SELECT count(*) FROM course_students WHERE deleted_at is null AND student_id = $1`
			if err := s.EurekaDB.QueryRow(ctx, csquery, pack.StudentPackage.StudentId).Scan(&countCourseStudent); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
			}

			//  will soft delete all old course of a student and upsert new courses
			// The testcase only have 1 course new in the request so result = 1
			if countCourseStudent != 2 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("upsert error expected to number of course_class %d, got %d", 2, countCourseStudent)
			}
			saquery := `SELECT count(*) FROM course_students_access_paths WHERE location_id = $2 AND student_id = $1 and course_id = $3`
			if err := s.EurekaDB.QueryRow(ctx, saquery, pack.StudentPackage.StudentId, pack.StudentPackage.Package.LocationId, pack.StudentPackage.Package.CourseId).Scan(&countCourseStudentAccessPath); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
			}
			if countCourseStudentAccessPath != 1 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("upsert error expected to number of course_students_access_paths %d, got %d", 2, countCourseStudent)
			}

		} else if !pack.StudentPackage.IsActive {
			query := `SELECT count(*) FROM course_students WHERE student_id = $1::TEXT AND deleted_at is not null`
			err := s.EurekaDB.QueryRow(ctx, query, pack.StudentPackage.StudentId).Scan(&countCourseStudent)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
			}

			// Testcase have 2 course but the request have 1 course need to be deleted
			// So result is 1
			if countCourseStudent != 1 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("deleted error expected to number of course_class %d, got %d", 1, countCourseStudent)
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
