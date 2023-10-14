package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ImportMasterDataService) ImportProductSetting(
	ctx context.Context,
	req *pb.ImportProductSettingRequest) (
	*pb.ImportProductSettingResponse,
	error,
) {
	errors := []*pb.ImportProductSettingResponse_ImportProductSettingError{}

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	headerTitles := []string{
		"product_id",
		"is_enrollment_required",
		"is_pausable",
		"is_added_to_enrollment_by_default",
		"is_operation_fee",
	}
	err = utils.ValidateCsvHeader(len(headerTitles), lines[0], headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	_ = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		hasError := false
		for i, line := range lines[1:] {
			productSetting, err := ProductSettingFromCsv(line)
			if err != nil {
				errors = append(errors, &pb.ImportProductSettingResponse_ImportProductSettingError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf("unable to parse product setting item: %s", err),
				})
				hasError = true
				continue
			}

			if _, err := s.ProductSettingRepo.GetByID(ctx, tx, productSetting.ProductID.String); err != nil {
				if err = s.ProductSettingRepo.Create(ctx, tx, &productSetting); err != nil {
					errors = append(errors, &pb.ImportProductSettingResponse_ImportProductSettingError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new product setting item: %s", err),
					})
					hasError = true
				}
			} else {
				if err := s.ProductSettingRepo.Update(ctx, tx, &productSetting); err != nil {
					errors = append(errors, &pb.ImportProductSettingResponse_ImportProductSettingError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update product setting item: %s", err),
					})
					hasError = true
				}
			}
		}
		if hasError {
			return fmt.Errorf("error importing product setting")
		}
		return nil
	})

	return &pb.ImportProductSettingResponse{
		Errors: errors,
	}, nil
}

func ProductSettingFromCsv(line []string) (productSetting entities.ProductSetting, err error) {
	const (
		ProductID = iota
		IsEnrollmentRequired
		IsPausable
		IsAddedToEnrollmentByDefault
		IsOperationFee
	)

	if err = multierr.Combine(
		utils.StringToFormatString("product_id", line[ProductID], false, productSetting.ProductID.Set),
		utils.StringToBool("is_enrollment_required", line[IsEnrollmentRequired], false, productSetting.IsEnrollmentRequired.Set),
		utils.StringToBool("is_pausable", line[IsPausable], false, productSetting.IsPausable.Set),
		utils.StringToBool("is_added_to_enrollment_by_default", line[IsAddedToEnrollmentByDefault], false, productSetting.IsAddedToEnrollmentByDefault.Set),
		utils.StringToBool("is_operation_fee", line[IsOperationFee], false, productSetting.IsOperationFee.Set),
	); err != nil {
		return productSetting, err
	}

	return productSetting, nil
}
