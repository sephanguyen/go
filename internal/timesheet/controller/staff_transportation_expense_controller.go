package controller

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StaffTransportationExpenseController struct {
	StaffTransportationExpenseService interface {
		UpsertConfig(ctx context.Context, staffId string, listStaffTransportExpenses *dto.ListStaffTransportationExpenses) error
	}
}

func (s *StaffTransportationExpenseController) UpsertStaffTransportationExpense(ctx context.Context, request *pb.UpsertStaffTransportationExpenseRequest) (*pb.UpsertStaffTransportationExpenseResponse, error) {

	err := validateStaffIDRequest(request.StaffId)
	if err != nil {
		return nil, err
	}

	listStaffTransportationExpensesReq := dto.NewListStaffTransportExpensesFromRPCRequest(request.StaffId, request.ListStaffTransportationExpenses)

	err = listStaffTransportationExpensesReq.ValidateUpsertInfo()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.StaffTransportationExpenseService.UpsertConfig(ctx, request.StaffId, &listStaffTransportationExpensesReq)
	if err != nil {
		return nil, err
	}

	return &pb.UpsertStaffTransportationExpenseResponse{Success: true}, nil
}

func validateStaffIDRequest(staffID string) error {
	if strings.TrimSpace(staffID) == "" {
		return status.Error(codes.InvalidArgument, "staff id must not be empty")
	}

	return nil
}
