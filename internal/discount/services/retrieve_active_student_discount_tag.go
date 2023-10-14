package services

import (
	"context"
	"errors"
	"strings"

	"github.com/manabie-com/backend/internal/discount/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *DiscountService) RetrieveActiveStudentDiscountTag(ctx context.Context, req *pb.RetrieveActiveStudentDiscountTagRequest) (*pb.RetrieveActiveStudentDiscountTagResponse, error) {
	// validate request
	if err := validateDiscountTagRequest(req); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	discountTagDetails := make([]*pb.RetrieveActiveStudentDiscountTagResponse_DiscountTagDetail, 0)
	response := &pb.RetrieveActiveStudentDiscountTagResponse{
		StudentId:          req.StudentId,
		DiscountTagDetails: discountTagDetails,
	}

	// retrieve unique active discount tag ids for student
	discountTagIDs, err := s.DiscountTagService.RetrieveActiveDiscountTagIDsByDateAndUserID(ctx, s.DB, req.DiscountDateRequest.AsTime(), req.StudentId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, discountTagID := range discountTagIDs {
		discountTag, err := s.DiscountTagService.RetrieveDiscountTagByDiscountTagID(ctx, s.DB, discountTagID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		discountTagDetail := &pb.RetrieveActiveStudentDiscountTagResponse_DiscountTagDetail{
			DiscountTagName: discountTag.DiscountTagName.String,
			DiscountTagId:   discountTag.DiscountTagID.String,
			Selectable:      discountTag.Selectable.Bool,
		}
		discountTagDetails = append(discountTagDetails, discountTagDetail)
	}

	response.DiscountTagDetails = discountTagDetails

	return response, nil
}

func validateDiscountTagRequest(req *pb.RetrieveActiveStudentDiscountTagRequest) error {
	if strings.TrimSpace(req.StudentId) == "" {
		return errors.New("student id should be required")
	}

	discountDateReq := req.DiscountDateRequest.AsTime()
	if utils.IsZeroTime(discountDateReq) {
		return errors.New("discount date request should be required")
	}

	return nil
}
