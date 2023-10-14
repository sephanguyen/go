package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/status"
)

func (s *suite) errorWith(ctx context.Context, err string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != err {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s, got %s status code, message: %s", err, stt.Code().String(), stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnWithError(ctx context.Context, err string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch err {
	case "null":
		ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanIDWithStatus(ctx, database.TextArray([]string{stepState.StudyPlanID}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
		}

		for _, each := range studyPlanItems {
			if each.Status.Status == pgtype.Null || each.Status.String == epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update school date")
			}
		}
	case "InvalidArgument":
		return s.errorWith(ctx, err)
	case "updated":
		ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanIDWithStatus(ctx, database.TextArray([]string{stepState.StudyPlanID}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
		}

		for _, each := range studyPlanItems {
			if each.Status.Status == pgtype.Null ||
				each.Status.String == epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE.String() ||
				each.Status.String == stepState.OldStudyPlanItemStatus.String() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update school date")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateStatusWith(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	var req *epb.UpdateStudyPlanItemsStatusRequest

	switch condition {
	case "valid_request":
		ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
		}

		ids := make([]string, 0)
		for _, item := range studyPlanItems {
			ids = append(ids, item.ID.String)
		}

		req = &epb.UpdateStudyPlanItemsStatusRequest{
			StudentId:           stepState.StudentID,
			StudyPlanItemIds:    ids,
			StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED,
		}
	case "missing_ids":
		req = &epb.UpdateStudyPlanItemsStatusRequest{
			StudyPlanItemIds: []string{},
			StudentId:        "student_id",
		}
	case "missing_student_id":
		req = &epb.UpdateStudyPlanItemsStatusRequest{
			StudyPlanItemIds: []string{"study-plan-item-id"},
			StudentId:        "",
		}
	case "missing_status":
		req = &epb.UpdateStudyPlanItemsStatusRequest{
			StudyPlanItemIds:    []string{"study-plan-item-id"},
			StudentId:           "student_id",
			StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_NONE,
		}
	case "update_request":
		ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
		}

		ids := make([]string, 0)
		for _, item := range studyPlanItems {
			ids = append(ids, item.ID.String)
		}

		req = &epb.UpdateStudyPlanItemsStatusRequest{
			StudentId:           stepState.StudentID,
			StudyPlanItemIds:    ids,
			StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED,
		}

		stepState.OldStudyPlanItemStatus = epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED
	}

	stepState.ResponseErr = try.Do(func(attempt int) (bool, error) {
		isRetryable := attempt < 3
		if _, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsStatus(ctx, req); err != nil {
			return isRetryable, err
		}
		return false, nil
	})

	if condition == "update_request" {
		req.StudyPlanItemStatus = epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE
		stepState.ResponseErr = try.Do(func(attempt int) (bool, error) {
			isRetryable := attempt < 3
			if _, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsStatus(ctx, req); err != nil {
				return isRetryable, err
			}
			return false, nil
		})
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudyPlanItemByStudyPlanIDWithStatus(ctx context.Context, studyPlanIDs pgtype.TextArray) (context.Context, []*entities.StudyPlanItem, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT study_plan_item_id, status FROM study_plan_items AS spi WHERE spi.study_plan_id = ANY($1::TEXT[]) AND deleted_at IS NULL;"

	rows, err := s.DB.Query(ctx, query, &studyPlanIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("something wrong: %w", err)
	}
	defer rows.Close()

	resp := make([]*entities.StudyPlanItem, 0)
	for rows.Next() {
		studyPlanItem := new(entities.StudyPlanItem)
		err := rows.Scan(&studyPlanItem.ID, &studyPlanItem.Status)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("something wrong: %w", err)
		}

		resp = append(resp, studyPlanItem)
	}

	return StepStateToContext(ctx, stepState), resp, nil
}
