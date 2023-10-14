package entryexitmgmt

import (
	"context"
	"fmt"
	"time"

	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) entryExitRecordIsUpdatedSuccessfully(ctx context.Context) (context.Context, error) {
	return checkResponseError(ctx)
}

func (s *suite) updatesTheRecordOfThisStudentIn(ctx context.Context, account, existingTouchEvent, timeZone string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return ctx, err
	}
	req, err := s.generateUpdateEntryExitRequest(ctx, timeZone, existingTouchEvent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to generate update entry exit record for student %s: %w", stepState.StudentID, err)
	}

	stepState.Request = req
	return s.sendEntryExitRequest(StepStateToContext(ctx, stepState), stepState.Request)
}

func (s *suite) updatesTheExistingRecordWithInvalidRequest(ctx context.Context, account, existingTouchEvent, invalidArg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, account)
	if err != nil {
		return ctx, err
	}

	entryExitRepo := s.StudentEntryExitRecordsRepo

	latestRecord, err := entryExitRepo.GetLatestRecordByID(ctx, s.EntryExitMgmtDBTrace, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	req := &eepb.UpdateEntryExitRequest{
		EntryExitPayload: &eepb.EntryExitPayload{
			StudentId:     stepState.StudentID,
			EntryDateTime: timestamppb.New(now.Add(-1 * time.Hour)),
			ExitDateTime:  nil,
			NotifyParents: stepState.NotifyParentRequest,
		},
		EntryexitId: latestRecord.ID.Int,
	}

	if existingTouchEvent == "exit" {
		req.EntryExitPayload.ExitDateTime = timestamppb.New(now)
	}

	if invalidArg == "cannot retrieve entry exit id in database" {
		req.EntryexitId = 0
	} else {
		ctx = generateInvalidRequest(ctx, req.EntryExitPayload, invalidArg)
	}

	stepState.Request = req
	return s.sendEntryExitRequest(StepStateToContext(ctx, stepState), stepState.Request)
}

func (s *suite) generateUpdateEntryExitRequest(ctx context.Context, timeZone, existingTouchEvent string) (*eepb.UpdateEntryExitRequest, error) {
	stepState := StepStateFromContext(ctx)
	entryExitRepo := s.StudentEntryExitRecordsRepo

	//  get the latest record inserted by integration test

	latestRecord, err := entryExitRepo.GetLatestRecordByID(ctx, s.EntryExitMgmtDBTrace, stepState.StudentID)
	if err != nil {
		return nil, err
	}

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	req := &eepb.UpdateEntryExitRequest{
		EntryExitPayload: &eepb.EntryExitPayload{
			StudentId:     stepState.StudentID,
			EntryDateTime: timestamppb.New(now.Add(-14 * time.Minute).In(location)),
			ExitDateTime:  nil,
			NotifyParents: stepState.NotifyParentRequest,
		},
		EntryexitId: latestRecord.ID.Int,
	}
	if existingTouchEvent == "exit" {
		req.EntryExitPayload.ExitDateTime = timestamppb.New(now.Add(-7 * time.Minute).In(location))
	}

	return req, nil
}
