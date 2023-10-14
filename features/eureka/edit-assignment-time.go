package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//nolint
func (s *suite) adminReupdateAssignmentTimeWith(ctx context.Context, updateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken

	req := &pb.EditAssignmentTimeRequest{
		StudyPlanItemIds: stepState.StudyPlanItemIDs,
		StudentId:        stepState.StudentID,
	}
	date := time.Now().AddDate(0, 0, 4)
	switch updateType {
	case "start":
		req.UpdateType = pb.UpdateType_UPDATE_START_DATE
		req.StartDate = timestamppb.New(date)
		stepState.NewStartDate = date
	case "end":
		req.UpdateType = pb.UpdateType_UPDATE_END_DATE
		req.EndDate = timestamppb.New(date.AddDate(0, 0, 4))
		stepState.NewEndDate = date.AddDate(0, 0, 4)
	case "start_end", "no_type":
		if updateType == "start_end" {
			req.UpdateType = pb.UpdateType_UPDATE_START_DATE_END_DATE
		}
		req.StartDate = timestamppb.New(date)
		req.EndDate = timestamppb.New(date.AddDate(0, 0, 4))
		stepState.NewStartDate = date
		stepState.NewEndDate = date.AddDate(0, 0, 4)
	}

	_, err := pb.NewAssignmentModifierServiceClient(s.Conn).EditAssignmentTime(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable edit assignment time: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) adminUpdateAssignmentTimeWith(ctx context.Context, updateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken

	ids := make([]string, 0)
	if err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(5 * time.Second)
		isRetryable := attempt < 10
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudentToDoItems(s.signedCtx(ctx), &pb.ListStudentToDoItemsRequest{
			StudentId: stepState.StudentID,
			CourseIds: []string{stepState.CourseID},
			Status:    pb.ToDoStatus_TO_DO_STATUS_ACTIVE,
		})
		if err != nil {
			return isRetryable, fmt.Errorf("unable fetch todo items: %w", err)
		}

		if len(resp.Items) == 0 {
			return isRetryable, fmt.Errorf("no item exists")
		}

		for _, item := range resp.Items {
			ids = append(ids, item.StudyPlanItem.StudyPlanItemId)
		}
		stepState.StudyPlanItemIDs = ids

		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &pb.EditAssignmentTimeRequest{
		StudyPlanItemIds: stepState.StudyPlanItemIDs,
		StudentId:        stepState.StudentID,
	}
	date := time.Now().AddDate(0, 0, 1)
	switch updateType {
	case "start":
		req.UpdateType = pb.UpdateType_UPDATE_START_DATE
		req.StartDate = timestamppb.New(date)
		stepState.NewStartDate = date
		stepState.OldStartDate = date
	case "end":
		req.UpdateType = pb.UpdateType_UPDATE_END_DATE
		req.EndDate = timestamppb.New(date.AddDate(0, 0, 2))
		stepState.NewEndDate = date.AddDate(0, 0, 2)
		stepState.OldEndDate = date.AddDate(0, 0, 2)
	case "start_end", "no_type":
		if updateType == "start_end" {
			req.UpdateType = pb.UpdateType_UPDATE_START_DATE_END_DATE
		}
		req.StartDate = timestamppb.New(date)
		req.EndDate = timestamppb.New(date.AddDate(0, 0, 2))
		stepState.NewStartDate = date
		stepState.NewEndDate = date.AddDate(0, 0, 2)
		stepState.OldStartDate = date
		stepState.OldEndDate = date.AddDate(0, 0, 2)
	case "missing_studyplanitem_ids":
		req.StudyPlanItemIds = []string{}
	case "missing_student_id":
		req.StudentId = ""
	case "invalid_date":
		req.UpdateType = pb.UpdateType_UPDATE_START_DATE_END_DATE
		req.EndDate = timestamppb.New(date)
		req.StartDate = timestamppb.New(date.AddDate(0, 0, 1))
	case "another_invalid_date":
		req.StartDate = timestamppb.New(date.AddDate(-2, 0, 0))
		req.EndDate = timestamppb.New(date.AddDate(-1, 0, 0))
	}

	_, err := pb.NewAssignmentModifierServiceClient(s.Conn).EditAssignmentTime(s.signedCtx(ctx), req)
	if err != nil {
		switch updateType {
		case "missing_studyplanitem_ids", "missing_student_id", "invalid_date", "another_invalid_date":
			stepState.ResponseErr = err
			return StepStateToContext(ctx, stepState), nil
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable edit assignment time: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) adminUpdateAssignmentTimeWithNullDataAnd(ctx context.Context, updateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken

	ids := make([]string, 0)
	if err := try.Do(func(attempt int) (bool, error) {
		isRetryable := attempt < 10
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudentToDoItems(s.signedCtx(ctx), &pb.ListStudentToDoItemsRequest{
			StudentId: stepState.StudentID,
			CourseIds: []string{stepState.CourseID},
			Status:    pb.ToDoStatus_TO_DO_STATUS_ACTIVE,
		})
		if err != nil {
			return isRetryable, fmt.Errorf("unable fetch todo items: %w", err)
		}

		if len(resp.Items) == 0 {
			return isRetryable, fmt.Errorf("no item exists")
		}

		for _, item := range resp.Items {
			ids = append(ids, item.StudyPlanItem.StudyPlanItemId)
		}
		stepState.StudyPlanItemIDs = ids

		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &pb.EditAssignmentTimeRequest{
		StudyPlanItemIds: stepState.StudyPlanItemIDs,
		StudentId:        stepState.StudentID,
	}
	switch updateType {
	case "start":
		req.UpdateType = pb.UpdateType_UPDATE_START_DATE
		req.StartDate = nil
	case "end":
		req.UpdateType = pb.UpdateType_UPDATE_END_DATE
		req.EndDate = nil
	case "start_end":
		req.UpdateType = pb.UpdateType_UPDATE_START_DATE_END_DATE
		req.StartDate = nil
		req.EndDate = nil
	case "no_type":
		req.StartDate = nil
		req.EndDate = nil
	}

	_, err := pb.NewAssignmentModifierServiceClient(s.Conn).EditAssignmentTime(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable edit assignment time: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) assignmentTimeWasUpdatedWithAccordingUpdate_type(ctx context.Context, updateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FindByIDs(ctx, s.DB, database.TextArray(stepState.StudyPlanItemIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable fetch study plan item: %w", err)
	}

	for _, item := range studyPlanItems {
		switch updateType {
		case "start":
			if !isEqual(item.StartDate.Time, stepState.NewStartDate) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "end":
			if !isEqual(item.EndDate.Time, stepState.NewEndDate) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "start_end", "no_type":
			if !isEqual(item.StartDate.Time, stepState.NewStartDate) && !isEqual(item.EndDate.Time, stepState.NewEndDate) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) assignmentTimeWasUpdatedWithNewDataAndAccordingUpdate_type(ctx context.Context, updateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FindByIDs(ctx, s.DB, database.TextArray(stepState.StudyPlanItemIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable fetch study plan item: %w", err)
	}

	for _, item := range studyPlanItems {
		switch updateType {
		case "start":
			if !isEqual(item.StartDate.Time, stepState.NewStartDate) || isEqual(item.StartDate.Time, stepState.OldStartDate) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "end":
			if !isEqual(item.EndDate.Time, stepState.NewEndDate) || isEqual(item.EndDate.Time, stepState.OldEndDate) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "start_end", "no_type":
			if (!isEqual(item.StartDate.Time, stepState.NewStartDate) && !isEqual(item.EndDate.Time, stepState.NewEndDate)) ||
				isEqual(item.StartDate.Time, stepState.OldStartDate) ||
				isEqual(item.EndDate.Time, stepState.OldEndDate) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) assignmentTimeWasUpdatedWithNullDataAndAccordingUpdate_type(ctx context.Context, updateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studyPlanItemRepo := &repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FindByIDs(ctx, s.DB, database.TextArray(stepState.StudyPlanItemIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable fetch study plan item: %w", err)
	}

	for _, item := range studyPlanItems {
		switch updateType {
		case "start":
			if item.StartDate.Status != pgtype.Null {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "end":
			if item.EndDate.Status != pgtype.Null {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		case "start_end", "no_type":
			if item.StartDate.Status != pgtype.Null || item.EndDate.Status != pgtype.Null {
				return StepStateToContext(ctx, stepState), fmt.Errorf("edit time fail with %s", updateType)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func isEqual(src, dest time.Time) bool {
	return src.Format(time.RFC3339) == dest.Format(time.RFC3339)
}
