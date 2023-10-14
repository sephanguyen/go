package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) returnError(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr.Error() != "rpc error: code = Unknown desc = empty study plan item ids" &&
		stepState.ResponseErr.Error() != "rpc error: code = Unknown desc = student id required" &&
		stepState.ResponseErr.Error() != "rpc error: code = Unknown desc = status is invalid" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong status code: %v", stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnSuccessfulWithUpdatedRecordWithSchoolDate(ctx context.Context, result string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
	}

	switch result {
	case "not null":
		for _, each := range studyPlanItems {
			// be careful about type timestamptz
			if each.SchoolDate.Status == pgtype.Status(pgtype.None) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update school date: have to not null")
			}
		}
	case "null":
		for _, each := range studyPlanItems {
			if each.SchoolDate.Status != pgtype.Null {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update school date: have to null")
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateSchoolDateWithMissingSchoolDate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
	}

	ids := make([]string, 0)
	for _, item := range studyPlanItems {
		ids = append(ids, item.ID.String)
	}

	if _, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsSchoolDate(ctx, &epb.UpdateStudyPlanItemsSchoolDateRequest{
		StudentId:        stepState.StudentID,
		StudyPlanItemIds: ids,
		SchoolDate:       timestamppb.New(time.Now()),
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update school date: %w", err)
	}

	if _, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsSchoolDate(ctx, &epb.UpdateStudyPlanItemsSchoolDateRequest{
		StudentId:        stepState.StudentID,
		StudyPlanItemIds: ids,
		SchoolDate:       nil,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update school date: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) updateSchoolDateWithMissingStudentId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken

	stepState.ResponseErr = try.Do(func(attempt int) (bool, error) {
		isRetryable := attempt < 3
		if _, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsSchoolDate(s.signedCtx(ctx), &epb.UpdateStudyPlanItemsSchoolDateRequest{
			StudyPlanItemIds: []string{"study-plan-item-id"},
			StudentId:        "",
		}); err != nil {
			return isRetryable, err
		}
		return false, nil
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateSchoolDateWithMissingStudyPlanItemIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	stepState.ResponseErr = try.Do(func(attempt int) (bool, error) {
		isRetryable := attempt < 3
		if _, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsSchoolDate(ctx, &epb.UpdateStudyPlanItemsSchoolDateRequest{
			StudyPlanItemIds: []string{},
		}); err != nil {
			return isRetryable, err
		}
		return false, nil
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateSchoolDateWithValidRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
	}

	ids := make([]string, 0)
	for _, item := range studyPlanItems {
		ids = append(ids, item.ID.String)
	}
	ctx = contextWithToken(s, ctx)

	if _, err := epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsSchoolDate(s.signedCtx(ctx), &epb.UpdateStudyPlanItemsSchoolDateRequest{
		StudentId:        stepState.StudentID,
		StudyPlanItemIds: ids,
		SchoolDate:       timestamppb.New(time.Now()),
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update school date: %w", err)
	}
	err = try.Do(func(attempt int) (bool, error) {
		ctx, studyPlanItems, err = s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
		if err != nil {
			return false, err
		}
		if len(studyPlanItems) > 0 {
			if studyPlanItems[0].CreatedAt != studyPlanItems[0].UpdatedAt {
				return false, nil
			}
		}
		return attempt < 5, nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the update study plan item school date not apply")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudyPlanItemByStudyPlanID(ctx context.Context, studyPlanIDs pgtype.TextArray) (context.Context, []*entities.StudyPlanItem, error) {
	stepState := StepStateFromContext(ctx)
	query := "SELECT study_plan_item_id, school_date FROM study_plan_items AS spi WHERE spi.study_plan_id = ANY($1::TEXT[]) AND deleted_at IS NULL;"

	rows, err := s.DB.Query(ctx, query, &studyPlanIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, fmt.Errorf("something wrong: %w", err)
	}
	defer rows.Close()

	resp := make([]*entities.StudyPlanItem, 0)
	for rows.Next() {
		studyPlanItem := &entities.StudyPlanItem{}
		database.AllNullEntity(studyPlanItem)
		err := rows.Scan(&studyPlanItem.ID, &studyPlanItem.SchoolDate)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, fmt.Errorf("something wrong: %w", err)
		}

		resp = append(resp, studyPlanItem)
	}
	if rows.Err() != nil {
		return StepStateToContext(ctx, stepState), resp, fmt.Errorf("row.Err:%w", err)
	}
	return StepStateToContext(ctx, stepState), resp, nil
}
