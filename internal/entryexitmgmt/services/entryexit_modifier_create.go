package services

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntryExitModifierService) CreateEntryExit(ctx context.Context, req *eepb.CreateEntryExitRequest) (*eepb.CreateEntryExitResponse, error) {
	student, err := s.validatePayloadAndGetStudent(ctx, req.EntryExitPayload)
	if err != nil {
		return nil, err
	}

	// generate entity from request
	studentEntryExitRecord, err := generateEntryExitCreateRecord(req.EntryExitPayload.StudentId, req.EntryExitPayload.EntryDateTime, req.EntryExitPayload.ExitDateTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// storing for database
	if err := s.StudentEntryExitRecordsRepo.Create(ctx, s.DB, studentEntryExitRecord); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &eepb.CreateEntryExitResponse{
		Successful:     true,
		Message:        "You have added a new record successfully!",
		ParentNotified: false,
	}

	response.ParentNotified, err = s.notifyManualTouch(ctx, req.EntryExitPayload, student, eepb.RecordType_CREATE_MANUAL)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return response, nil
}

func (s *EntryExitModifierService) notifyManualTouch(ctx context.Context, payload *eepb.EntryExitPayload, student *entities.Student, recordType eepb.RecordType) (bool, error) {
	if !payload.NotifyParents {
		return false, nil
	}

	return s.notifyParents(ctx, student, &EntryExitNotifyDetails{
		Student:    student,
		TouchEvent: eepb.TouchEvent_TOUCH_MANUAL_RECORD,
		TouchTime:  time.Time{},
		RecordType: recordType,
	})
}

func (s *EntryExitModifierService) validatePayloadAndGetStudent(ctx context.Context, payload *eepb.EntryExitPayload) (*entities.Student, error) {
	// validating request from client
	if err := validateRequest(payload); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	student, err := s.getStudentWithEEAccess(ctx, payload.StudentId, true)
	if err != nil {
		return nil, err
	}

	return student, nil
}
