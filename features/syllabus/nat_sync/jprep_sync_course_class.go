package nat_sync

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

func (s *Suite) NatSendARequestSyncCourseClass(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	action := []npb.ActionKind{npb.ActionKind_ACTION_KIND_UPSERTED, npb.ActionKind_ACTION_KIND_DELETED, npb.ActionKind_ACTION_KIND_UPSERTED}
	classes := []*npb.EventMasterRegistration_Class{}

	for idx, cs := range stepState.CourseClass {
		classID, err := strconv.ParseUint(cs.ClassID.String, 10, 64)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		class := &npb.EventMasterRegistration_Class{ClassId: classID, CourseId: cs.CourseID.String, ActionKind: action[idx]}
		classes = append(classes, class)
	}

	req := &npb.EventMasterRegistration{Classes: classes}
	stepState.Request = req
	courseClassService := &services.CourseClassService{
		DB:              s.EurekaDBTrace,
		CourseClassRepo: &repositories.CourseClassRepo{},
	}
	if err := courseClassService.SyncCourseClass(ctx, req); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) StoreCorrectResultFromSyncCourseClassRequest(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := stepState.Request.(*npb.EventMasterRegistration)

	for _, classReq := range req.Classes {
		var count int

		if classReq.ActionKind == npb.ActionKind_ACTION_KIND_UPSERTED {
			query := `SELECT count(*) FROM course_classes WHERE course_id = $1 AND class_id = $2`
			classID := fmt.Sprintf("%v", classReq.ClassId)
			err := s.EurekaDB.QueryRow(ctx, query, classReq.CourseId, classID).Scan(&count)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
			}

			if count != 1 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of course_class %d, got %d", 1, count)
			}
		} else if classReq.ActionKind == npb.ActionKind_ACTION_KIND_DELETED {
			classID := fmt.Sprintf("%v", classReq.ClassId)
			query := `SELECT count(*) FROM course_classes WHERE course_id = $1 AND class_id = $2::TEXT AND deleted_at is not null`
			err := s.EurekaDB.QueryRow(ctx, query, classReq.CourseId, classID).Scan(&count)
			if err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("error can not get result: %w", err)
			}

			if count != 1 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected to number of course_class %d, got %d", 1, count)
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
