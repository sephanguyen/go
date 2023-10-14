package entryexitmgmt

import (
	"context"
	"fmt"

	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
)

func (s *suite) createsRecordOfThisStudentIn(ctx context.Context, account, entryexit, timeZone string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return ctx, err
	}
	req, err := generateCreateEntryExitRequest(ctx, entryexit, timeZone)
	if err != nil {
		return ctx, fmt.Errorf("unable to generate create entry exit record for student %s: %w", stepState.StudentID, err)
	}

	stepState.Request = req
	return s.sendEntryExitRequest(StepStateToContext(ctx, stepState), stepState.Request)
}

func (s *suite) newEntryExitRecordIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	return checkResponseError(ctx)
}

func (s *suite) receivesStatusCode(ctx context.Context, expectedCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.CommonSuite.ReturnsStatusCode(StepStateToContext(ctx, stepState), expectedCode)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createsInvalidRequest(ctx context.Context, account, invalidArg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return ctx, err
	}
	req := &eepb.CreateEntryExitRequest{
		EntryExitPayload: &eepb.EntryExitPayload{
			StudentId:     stepState.StudentID,
			EntryDateTime: nil,
			ExitDateTime:  nil,
			NotifyParents: stepState.NotifyParentRequest,
		},
	}
	ctx = generateInvalidRequest(ctx, req.EntryExitPayload, invalidArg)
	stepState.Request = req
	return s.sendEntryExitRequest(StepStateToContext(ctx, stepState), stepState.Request)
}

func (s *suite) notifyParentsCheckbox(ctx context.Context, account, notifStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.NotifyParentRequest = notifStatus == "checked"

	return StepStateToContext(ctx, stepState), nil
}
