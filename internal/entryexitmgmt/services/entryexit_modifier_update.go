package services

import (
	"context"
	"time"

	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *EntryExitModifierService) UpdateEntryExit(ctx context.Context, req *eepb.UpdateEntryExitRequest) (*eepb.UpdateEntryExitResponse, error) {
	// validating request from client
	if int(req.EntryexitId) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "student entry exit id must be valid")
	}

	student, err := s.validatePayloadAndGetStudent(ctx, req.EntryExitPayload)
	if err != nil {
		return nil, err
	}

	// generate entity from update request
	exitTime := time.Time{}
	if req.EntryExitPayload.ExitDateTime != nil {
		exitTime = req.EntryExitPayload.ExitDateTime.AsTime()
	}

	studentEntryExitRecord, err := generateEntryExitUpdateRecord(
		req.EntryexitId,
		timestamppb.New(req.EntryExitPayload.EntryDateTime.AsTime()),
		timestamppb.New(exitTime),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// update record
	if err := s.StudentEntryExitRecordsRepo.Update(ctx, s.DB, studentEntryExitRecord); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &eepb.UpdateEntryExitResponse{
		Successful:     true,
		ParentNotified: false,
	}

	response.ParentNotified, err = s.notifyManualTouch(ctx, req.EntryExitPayload, student, eepb.RecordType_UPDATE_MANUAL)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return response, nil
}
