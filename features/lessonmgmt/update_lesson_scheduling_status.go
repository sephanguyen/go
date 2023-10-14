package lessonmgmt

import (
	"context"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) userChangedStatusTo(ctx context.Context, status, sType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var schedulingStatus cpb.LessonSchedulingStatus
	if status == "draft" {
		schedulingStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT
	}
	var savingType lpb.SavingType
	if sType == "only this" {
		savingType = lpb.SavingType_THIS_ONE
	} else if sType == "this and following" {
		savingType = lpb.SavingType_THIS_AND_FOLLOWING
	}
	stepState.SavingType = savingType
	updatedReq := &lpb.UpdateLessonSchedulingStatusRequest{
		LessonId:         stepState.CurrentLessonID,
		SchedulingStatus: schedulingStatus,
		SavingType:       savingType,
	}
	stepState.Response, stepState.ResponseErr = lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).UpdateLessonSchedulingStatus(s.CommonSuite.SignedCtx(ctx), updatedReq)
	stepState.Request = updatedReq
	return StepStateToContext(ctx, stepState), nil
}
