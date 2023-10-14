package entryexitmgmt

import (
	"context"

	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
)

func (s *suite) deletesThatRecordOfThisStudent(ctx context.Context, entryexit string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	entryExitRepo := s.StudentEntryExitRecordsRepo

	latestRecord, err := entryExitRepo.GetLatestRecordByID(ctx, s.EntryExitMgmtDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &eepb.DeleteEntryExitRequest{
		EntryexitId: latestRecord.ID.Int,
		StudentId:   stepState.StudentID,
	}

	ctx, err = s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return ctx, err
	}

	return s.sendEntryExitRequest(StepStateToContext(ctx, stepState), req)
}
