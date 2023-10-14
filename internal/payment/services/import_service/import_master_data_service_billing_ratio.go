package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"

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

func (s *ImportMasterDataService) ImportBillingRatio(ctx context.Context, req *pb.ImportBillingRatioRequest) (*pb.ImportBillingRatioResponse, error) {
	errors := []*pb.ImportBillingRatioResponse_ImportBillingRatioError{}

	r := csv.NewReader(bytes.NewReader(req.Payload))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(lines) < 2 {
		return nil, status.Error(codes.InvalidArgument, constant.NoDataInCsvFile)
	}

	headerTitles := []string{
		"billing_ratio_id",
		"start_date",
		"end_date",
		"billing_schedule_period_id",
		"billing_ratio_numerator",
		"billing_ratio_denominator",
		"is_archived",
	}

	header := lines[0]
	if len(header) != len(headerTitles) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - number of columns should be %d", len(headerTitles)))
	}

	err = utils.ValidateCsvHeader(len(headerTitles), header, headerTitles)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("csv file invalid format - %s", err.Error()))
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// first line is header
		for i, line := range lines[1:] {
			billingRatio, err := CreateBillingRatioEntityFromCsv(line, headerTitles)
			if err != nil {
				errors = append(errors, &pb.ImportBillingRatioResponse_ImportBillingRatioError{
					RowNumber: int32(i) + 2, // i = 0 <=> line number 2 in csv file
					Error:     fmt.Sprintf(constant.UnableToParseBillingRatioItem, err),
				})
				continue
			}
			if billingRatio.BillingRatioID.Get() == nil {
				err := s.BillingRatioRepo.Create(ctx, tx, billingRatio)
				if err != nil {
					errors = append(errors, &pb.ImportBillingRatioResponse_ImportBillingRatioError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to create new billing ratio item: %s", err),
					})
				}
			} else {
				err := s.BillingRatioRepo.Update(ctx, tx, billingRatio)
				if err != nil {
					errors = append(errors, &pb.ImportBillingRatioResponse_ImportBillingRatioError{
						RowNumber: int32(i) + 2,
						Error:     fmt.Sprintf("unable to update billing ratio item: %s", err),
					})
				}
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf(errors[0].Error)
		}
		return nil
	})
	if err != nil {
		log.Printf("Error when importing billing ratio: %s", err.Error())
	}
	return &pb.ImportBillingRatioResponse{
		Errors: errors,
	}, nil
}

func CreateBillingRatioEntityFromCsv(line []string, columnNames []string) (*entities.BillingRatio, error) {
	const (
		BillingRatioID = iota
		StartDate
		EndDate
		BillingSchedulePeriodID
		BillingRatioNumerator
		BillingRatioDenominator
		IsArchived
	)

	mandatory := []int{StartDate, EndDate, BillingSchedulePeriodID, BillingRatioNumerator, BillingRatioDenominator, IsArchived}

	areMandatoryDataPresent, colPosition := checkMandatoryColumnAndGetIndex(line, mandatory)
	if !areMandatoryDataPresent {
		return nil, fmt.Errorf("missing mandatory data: %v", columnNames[colPosition])
	}

	billingRatio := &entities.BillingRatio{}

	if err := multierr.Combine(
		utils.StringToFormatString("billing_ratio_id", line[BillingRatioID], true, billingRatio.BillingRatioID.Set),
		utils.StringToDate("start_date", line[StartDate], false, billingRatio.StartDate.Set),
		utils.StringToDate("end_date", line[EndDate], false, billingRatio.EndDate.Set),
		utils.StringToFormatString("billing_schedule_period_id", line[BillingSchedulePeriodID], false, billingRatio.BillingSchedulePeriodID.Set),
		utils.StringToInt("billing_ratio_numerator", line[BillingRatioNumerator], false, billingRatio.BillingRatioNumerator.Set),
		utils.StringToInt("billing_ratio_denominator", line[BillingRatioDenominator], false, billingRatio.BillingRatioDenominator.Set),
		utils.StringToBool("is_archived", line[IsArchived], false, billingRatio.IsArchived.Set),
	); err != nil {
		return nil, err
	}

	if billingRatio.BillingRatioNumerator.Int < 0 {
		return nil, fmt.Errorf("billing_ratio_numerator should be >= 0")
	}
	if billingRatio.BillingRatioDenominator.Int < 1 {
		return nil, fmt.Errorf("billing_ratio_denominator should be >= 1")
	}
	if billingRatio.EndDate.Time.Before(billingRatio.StartDate.Time) {
		return nil, fmt.Errorf("start_date should be before end_date")
	}

	return billingRatio, nil
}
