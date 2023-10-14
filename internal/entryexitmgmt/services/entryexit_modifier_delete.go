package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntryExitModifierService) DeleteEntryExit(ctx context.Context, req *eepb.DeleteEntryExitRequest) (*eepb.DeleteEntryExitResponse, error) {
	if err := validateDeleteEntryExitRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err := s.getStudentWithEEAccess(ctx, req.StudentId, false)
	if err != nil {
		return nil, err
	}

	if err := s.StudentEntryExitRecordsRepo.SoftDeleteByID(ctx, s.DB, database.Int4(req.EntryexitId)); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &eepb.DeleteEntryExitResponse{
		Successful: true,
	}, nil
}

func validateDeleteEntryExitRequest(req *eepb.DeleteEntryExitRequest) error {
	if strings.TrimSpace(req.StudentId) == "" {
		return fmt.Errorf("student id cannot be empty")
	}

	if req.EntryexitId == 0 {
		return fmt.Errorf("invalid entry exit id")
	}

	return nil
}
