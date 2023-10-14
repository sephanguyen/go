package nat_sync

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

func (s *Suite) NatSendARequestSyncCourseStudentV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	req := &npb.EventSyncStudentPackage{StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
		{
			StudentId:  stepState.StudentIDs[1],
			Packages:   []*npb.EventSyncStudentPackage_Package{{CourseIds: []string{"course-3"}}},
			ActionKind: npb.ActionKind_ACTION_KIND_DELETED},
		{
			StudentId:  stepState.StudentIDs[0],
			Packages:   []*npb.EventSyncStudentPackage_Package{{CourseIds: []string{"course-4", "course-10"}}},
			ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
		},
	}}
	stepState.Request = req

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) StoreCorrectResultFromSyncCourseStudentRequest(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := stepState.Request.(*npb.EventSyncStudentPackage)
	repo := &repositories.CourseStudentRepo{}
	if err := services.ProcessSyncCourseStudent(ctx, s.EurekaDB, req, repo); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not processSyncCourseStudent: %w", err)
	}

	for _, pack := range req.StudentPackages {
		var count int

		if pack.ActionKind == npb.ActionKind_ACTION_KIND_UPSERTED {
			query := `SELECT count(*) FROM course_students WHERE deleted_at is null AND student_id = $1`
			err := s.EurekaDB.QueryRow(ctx, query, pack.StudentId).Scan(&count)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
			}

			// Upsert Actionkind will soft delete all old course of a student and upsert new courses
			// The testcase only have 1 course new in the request so result = 1
			if count != 2 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("upsert error expected to number of course_class %d, got %d", 2, count)
			}
		} else if pack.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
			query := `SELECT count(*) FROM course_students WHERE student_id = $1::TEXT AND deleted_at is not null`
			err := s.EurekaDB.QueryRow(ctx, query, pack.StudentId).Scan(&count)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
			}

			// Testcase have 2 course but the request have 1 course need to be deleted
			// So result is 1
			if count != 1 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("deleted error expected to number of course_class %d, got %d", 1, count)
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
