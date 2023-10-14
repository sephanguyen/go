package services

import (
	"context"
	"strings"

	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *EntryExitModifierService) RetrieveStudentQRCode(ctx context.Context, req *eepb.RetrieveStudentQRCodeRequest) (*eepb.RetrieveStudentQRCodeResponse, error) {
	if strings.TrimSpace(req.StudentId) == "" {
		return nil, status.Error(codes.InvalidArgument, "student id cannot be empty")
	}
	// check if student is existing
	student, err := s.StudentRepo.FindByID(ctx, s.DB, req.StudentId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// retrieve student qr code
	qrURL, err := s.Generate(ctx, student.ID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &eepb.RetrieveStudentQRCodeResponse{
		QrUrl: qrURL,
	}, nil
}
